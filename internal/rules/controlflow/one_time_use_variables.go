package controlflow

import (
	"strings"
	"unicode/utf8"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/phpdoc"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// oneTimeUseVariables reports a variable assigned and consumed once in the
// very next return / throw / destructuring statement.
type oneTimeUseVariables struct{}

func init() { register(oneTimeUseVariables{}) }

func (oneTimeUseVariables) ID() string { return "OneTimeUseVariables" }

func (oneTimeUseVariables) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KReturn, syntax.KExprStmt}
}

func (oneTimeUseVariables) Check(ctx *analysis.Context, n syntax.Node) {
	consumer := n.(syntax.Stmt)
	var arg syntax.Expr
	switch s := n.(type) {
	case *syntax.Return: // D1
		if !ctx.Bool("ANALYZE_RETURN_STATEMENTS") || s.Expr == nil || otuReturnsByRef(s) {
			return
		}
		arg = s.Expr
	case *syntax.ExprStmt:
		switch e := s.Expr.(type) {
		case *syntax.Throw: // D2
			if !ctx.Bool("ANALYZE_THROW_STATEMENTS") {
				return
			}
			arg = e.Expr
		case *syntax.Assign: // D3
			if !ctx.Bool("ANALYZE_ARRAY_DESTRUCTURING") || e.Op.Kind != syntax.TEqual {
				return
			}
			switch t := e.Var.(type) {
			case *syntax.List:
			case *syntax.Array:
				if !t.Short {
					return
				}
			default:
				return
			}
			arg = e.Value
		}
	}
	if arg == nil {
		return
	}
	subject, member := otuSubject(syntax.UnwrapParens(arg)) // D4
	if subject == nil {
		return
	}
	name := subject.Name

	prev, ok := util.PrevStmtNoDoc(ctx.File, consumer) // D5
	if !ok {
		return
	}
	ps, ok := prev.(*syntax.ExprStmt)
	if !ok {
		return
	}
	as, ok := ps.Expr.(*syntax.Assign)
	if !ok || as.Op.Kind != syntax.TEqual || as.Value == nil {
		return
	}
	lhs, ok := as.Var.(*syntax.Variable)
	if !ok || lhs.NameExpr != nil || lhs.Name != name {
		return
	}
	val := syntax.UnwrapParens(as.Value)

	if s := syntax.EnclosingFuncLike(consumer); s != nil { // D6
		for _, p := range syntax.FuncLikeParams(s) {
			if p.ByRef && p.Var != nil && p.Var.Name == name {
				return
			}
		}
		if c, ok := s.(*syntax.Closure); ok {
			for _, u := range c.Uses {
				if u.ByRef && u.Var != nil && u.Var.Name == name {
					return
				}
			}
		}
	}

	if !ctx.Bool("ALLOW_LONG_STATEMENTS") && utf8.RuneCountInString(ctx.Text(as)) > 80 { // D7
		return
	}

	// D8: nil scope = file-level code, counted like a function body.
	reads, writes := 0, 0
	scope := syntax.EnclosingFuncLike(as)
	for _, a := range util.VarAccesses(ctx.File, scope, name) {
		if !otuReachable(ctx.File, a.Var, scope) {
			continue
		}
		if a.Write {
			writes++
			if w, ok := a.By.(*syntax.Assign); ok && w.Op.Kind == syntax.TEqual && otuVarAnnotated(ctx, w, name) {
				return
			}
		} else {
			reads++
		}
		if reads > 1 || writes > 1 {
			return
		}
	}

	if _, isNew := val.(*syntax.New); isNew && ctx.PHP < phpver.PHP54 { // D9
		return
	}

	ctx.ReportNode(lhs, "Variable $"+name+" is used only once; inline its value.", analysis.Fix{
		Title: "Inline the variable",
		Edits: func() []analysis.TextEdit {
			var edits []analysis.TextEdit
			if doc, ok := util.DocCommentBefore(ctx.File, ps.Span().Start); ok { // F1
				edits = append(edits, analysis.TextEdit{Span: syntax.Span{Start: doc.Start, End: doc.End}})
			}
			edits = append(edits, analysis.TextEdit{Span: util.WithTrailingWhitespace(ctx.File, ps.Span())}) // F2, F4
			text := ctx.Text(val)
			if member && otuNeedsParens(val) {
				text = "(" + text + ")"
			}
			return append(edits, analysis.TextEdit{Span: subject.Span(), NewText: text}) // F3
		},
	})
}

// otuReachable reports whether n can be reached from the entry of scope
// (nil = file level, where the top-level statement list is checked too).
func otuReachable(f *syntax.File, n, scope syntax.Node) bool {
	if !util.Reachable(n, scope) {
		return false
	}
	if scope != nil {
		return true
	}
	top := n
	for top.Parent() != nil {
		top = top.Parent()
	}
	for _, s := range f.Stmts {
		if syntax.Node(s) == top {
			break
		}
		if util.Terminates(s) {
			return false
		}
	}
	return true
}

// otuReturnsByRef reports whether the nearest enclosing named function or
// method returns by reference (closures are skipped).
func otuReturnsByRef(n syntax.Node) bool {
	for p := n.Parent(); p != nil; p = p.Parent() {
		switch f := p.(type) {
		case *syntax.Function:
			return f.ByRef
		case *syntax.Method:
			return f.ByRef
		}
	}
	return false
}

// otuSubject returns the consumed variable: the argument itself, or the
// object of a single-level `->` access; member is true in the latter case.
func otuSubject(e syntax.Expr) (*syntax.Variable, bool) {
	plain := func(x syntax.Expr) *syntax.Variable {
		if v, ok := x.(*syntax.Variable); ok && v.NameExpr == nil && v.Name != "" {
			return v
		}
		return nil
	}
	switch x := e.(type) {
	case *syntax.Variable:
		return plain(x), false
	case *syntax.PropertyFetch:
		if v := plain(x.Var); v != nil {
			return v, true
		}
	case *syntax.MethodCall:
		if v := plain(x.Var); v != nil {
			return v, true
		}
	}
	return nil, false
}

// otuVarAnnotated reports whether the statement of assignment as is
// immediately preceded by a comment with exactly one @var tag naming $name.
func otuVarAnnotated(ctx *analysis.Context, as *syntax.Assign, name string) bool {
	stmt, ok := as.Parent().(*syntax.ExprStmt)
	if !ok {
		return false
	}
	// a statement is always preceded by at least the open tag
	i := util.TokenIndex(ctx.File, stmt.Span().Start) - 1
	for ctx.File.Tokens[i].Kind == syntax.TWhitespace {
		i--
	}
	t := ctx.File.Tokens[i]
	text := ctx.SpanText(syntax.Span{Start: t.Start, End: t.End})
	if t.Kind != syntax.TDocComment && !(t.Kind == syntax.TComment && strings.HasPrefix(text, "/*")) {
		return false
	}
	tags := phpdoc.Parse(text).All("var")
	if len(tags) != 1 {
		return false
	}
	typ, rest := phpdoc.SplitType(tags[0].Text)
	v := phpdoc.VarName(rest)
	if strings.HasPrefix(typ, "$") {
		v = phpdoc.VarName(typ)
	}
	return v == name
}

// otuNeedsParens reports whether val must be parenthesised to become the
// object of a `->` access.
func otuNeedsParens(val syntax.Expr) bool {
	switch val.(type) {
	case *syntax.New, *syntax.Clone, *syntax.Unary, *syntax.Closure, *syntax.IncDec:
		return true
	}
	return util.NeedsParensAsUnaryOperand(val)
}
