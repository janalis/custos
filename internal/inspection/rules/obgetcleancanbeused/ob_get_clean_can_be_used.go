package obgetcleancanbeused

import (
	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// obGetCleanCanBeUsed reports `ob_get_contents()` immediately followed by an
// `ob_end_clean();` statement.
type obGetCleanCanBeUsed struct{}

func (obGetCleanCanBeUsed) ID() string               { return "ObGetCleanCanBeUsed" }
func (obGetCleanCanBeUsed) Semantic()                {}
func (obGetCleanCanBeUsed) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (obGetCleanCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	end, ok := astquery.IsFuncNamedFold(n, "ob_end_clean") // D1
	if !ok || isCallableRef(end) {
		return
	}
	stmt, ok := end.Parent().(*syntax.ExprStmt)
	if !ok || stmt.Expr != syntax.Expr(end) {
		return
	}
	prev, ok := astquery.PrevStmtNoDoc(ctx.File, stmt) // D2
	if !ok {
		return
	}
	ps, ok := prev.(*syntax.ExprStmt)
	if !ok {
		return
	}
	var get *syntax.FuncCall // D3
	count := 0
	syntax.Inspect(ps, func(m syntax.Node) bool {
		switch x := m.(type) {
		case *syntax.Closure, *syntax.ArrowFunction, *syntax.ClassLike:
			return false // D5: other scopes, not run as part of the statement
		case *syntax.FuncCall:
			if c, ok := astquery.IsFuncNamedFold(x, "ob_get_contents"); ok && !isCallableRef(c) {
				if get == nil {
					get = c
				}
				count++
			}
		}
		return true
	})
	if get == nil || count > 1 || !semanticquery.ResolvesToGlobalFunction(ctx.Names(), ctx.Index(), ctx.PHP, end, "ob_end_clean") || !semanticquery.ResolvesToGlobalFunction(ctx.Names(), ctx.Index(), ctx.PHP, get, "ob_get_contents") { // D4
		return
	}
	always, inert := evaluation(ctx, get, ps)
	if !always { // D6
		return
	}
	const msg = "Use ob_get_clean() instead of ob_get_contents() + ob_end_clean()."
	if !inert { // F3
		ctx.ReportNode(get, msg)
		return
	}
	name := get.Name.(*syntax.Name)
	ctx.ReportNode(get, msg, diagnostic.Fix{
		Title: "Use ob_get_clean()",
		Edits: func() []diagnostic.TextEdit {
			// The removed statement takes the whitespace before it along
			// (no blank line with trailing spaces is left).
			del := stmt.Span()
			if ws, ok := astquery.TokenBefore(ctx.File, del.Start); ok && ws.Kind == syntax.TWhitespace {
				del.Start = ws.Start
			}
			return []diagnostic.TextEdit{
				{Span: name.Span(), NewText: semanticquery.QualifiedBuiltinFor(ctx, "ob_get_clean", get)},
				{Span: del, NewText: ""},
			}
		},
	})
}

// isCallableRef reports whether c is the first-class callable `f(...)`, which
// creates a closure instead of calling f.
func isCallableRef(c *syntax.FuncCall) bool {
	if len(c.Args.Args) != 1 {
		return false
	}
	_, ok := c.Args.Args[0].(*syntax.VariadicPlaceholder)
	return ok
}

// evaluation classifies how get is evaluated inside the statement ps. always
// is false when get sits in a part that may be skipped (D6). inert is true
// when nothing that could write to or observe the buffer runs between get
// and the end of the statement (F3).
func evaluation(ctx *analysis.Context, get *syntax.FuncCall, ps *syntax.ExprStmt) (always, inert bool) {
	inert = true
	for c, p := syntax.Node(get), get.Parent(); p != syntax.Node(ps); c, p = p, p.Parent() {
		if skippedOnNull(p, c) {
			return false, false
		}
		switch x := p.(type) {
		case *syntax.Paren, *syntax.Arg, *syntax.ArgList:
		case *syntax.Unary:
			inert = inert && x.Op.Kind.IsCast()
		case *syntax.FuncCall:
			inert = inert && inertWrapper(ctx, x)
		case *syntax.Assign:
			if x.Op.Kind == syntax.TCoalesceEqual && syntax.Node(x.Value) == c {
				return false, false
			}
			// A compound operator reads and converts the target after the
			// call; a property or element target may run a setter.
			v, simple := x.Var.(*syntax.Variable)
			inert = inert && simple && v.NameExpr == nil && x.Op.Kind == syntax.TEqual
		case *syntax.Ternary:
			if syntax.Node(x.Cond) != c {
				return false, false
			}
			inert = false
		case *syntax.Binary:
			switch x.Op.Kind {
			case syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr, syntax.TCoalesce:
				if syntax.Node(x.Right) == c {
					return false, false
				}
			}
			inert = false
		case *syntax.MatchArm:
			return false, false
		default:
			inert = false
		}
	}
	return true, inert
}

// receiver returns the expression a member access or call p is applied to and
// whether that step is nullsafe.
func receiver(p syntax.Node) (syntax.Expr, bool) {
	switch x := p.(type) {
	case *syntax.MethodCall:
		return x.Var, x.NullSafe
	case *syntax.PropertyFetch:
		return x.Var, x.NullSafe
	case *syntax.ArrayDimFetch:
		return x.Var, false
	}
	return nil, false
}

// skippedOnNull reports whether c, a child of p other than its receiver, is
// not evaluated when a nullsafe step of p's receiver chain yields null.
func skippedOnNull(p, c syntax.Node) bool {
	e, nullsafe := receiver(p)
	if e == nil || syntax.Node(e) == c {
		return false
	}
	for ; e != nil; e, nullsafe = receiver(e) {
		if nullsafe {
			return true
		}
	}
	return false
}

// inertWrapper reports whether call is a built-in that only computes a value
// from its single string argument, without output or buffer access.
func inertWrapper(ctx *analysis.Context, call *syntax.FuncCall) bool {
	if len(call.Args.Args) != 1 {
		return false
	}
	if a, ok := call.Args.Args[0].(*syntax.Arg); !ok || a.Unpack {
		return false
	}
	for _, name := range [...]string{"strlen", "trim", "ltrim", "rtrim"} {
		if _, ok := astquery.IsFuncNamedFold(call, name); ok {
			return semanticquery.ResolvesToGlobalFunction(ctx.Names(), ctx.Index(), ctx.PHP, call, name)
		}
	}
	return false
}
