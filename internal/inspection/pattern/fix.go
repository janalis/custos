package pattern

import (
	"fmt"
	"slices"
	"strings"

	"github.com/janalis/custos/internal/diagnostic"
	"github.com/janalis/custos/internal/inspection/astquery"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
)

// template is a compiled fix: a snippet whose metavariables are replaced
// by what they bound. An empty template deletes the matched statement.
type template struct {
	snip  *snippet // nil for an empty template
	slots []slot   // metavariable occurrences, in source order
	vars  []string
}

// slot is one metavariable occurrence in a template.
type slot struct {
	name   string
	span   syntax.Span // in the template text
	spread bool        // `...$REST` argument: span covers the `...`
	drop   syntax.Span // what an empty `...$REST` removes, separator included
	stmt   syntax.Span // for `$S;`: the whole statement, which a bound statement replaces
	isStmt bool        // the slot is a statement of its own (`$S;`)
	slot   slotLevel   // what may stand here without parentheses
}

func parseTemplate(text string) (*template, error) {
	t := &template{}
	if strings.TrimSpace(text) == "" {
		return t, nil
	}
	s, err := parseSnippet(text)
	if err != nil {
		return nil, err
	}
	t.snip, t.vars = s, s.vars
	base := uint32(len(snippetPrefix))
	rel := func(sp syntax.Span) syntax.Span { return syntax.Span{Start: sp.Start - base, End: sp.End - base} }
	var inString bool
	for _, root := range s.nodes() {
		syntax.Inspect(root, func(n syntax.Node) bool {
			name := metaVarName(n)
			if name == "" {
				return true
			}
			for p := n.Parent(); p != nil; p = p.Parent() {
				inString = inString || p.Kind() == syntax.KInterpolatedString
			}
			sp := n.Span()
			sl := slot{name: name, slot: level(n)}
			if a, ok := n.Parent().(*syntax.Arg); ok && a.Unpack && a.Value == n {
				sp, sl.spread, sl.drop = a.Span(), true, rel(dropSpan(a))
			}
			if es, ok := n.Parent().(*syntax.ExprStmt); ok {
				// The snippet's text ends before the `;` of a lone expression.
				stmt := rel(es.Span())
				stmt.End = min(stmt.End, uint32(len(s.text)))
				sl.stmt, sl.isStmt = stmt, true
			}
			sl.span = rel(sp)
			t.slots = append(t.slots, sl)
			return false
		})
	}
	if inString {
		return nil, fmt.Errorf("a metavariable inside a string would not stand for the code it matched")
	}
	return t, nil
}

// dropSpan returns what removing argument a removes: a with the separator
// before it, else the one after it.
func dropSpan(a *syntax.Arg) syntax.Span {
	args := a.Parent().(*syntax.ArgList).Args
	i := slices.Index(args, syntax.Expr(a))
	switch {
	case i > 0:
		return syntax.Span{Start: args[i-1].Span().End, End: a.Span().End}
	case i < len(args)-1:
		return syntax.Span{Start: a.Span().Start, End: args[i+1].Span().Start}
	}
	return a.Span()
}

// slotLevel tells what expression may stand at a position without
// parentheses.
type slotLevel uint8

const (
	anyExpr    slotLevel = iota // delimited: an argument, a condition...
	assignable                  // an assigned value: all but and/or/xor
	operand                     // an operand: only atomic expressions
	className                   // the class of new: a name, variable or fetch
)

// level returns what may stand at n's position: anything in an argument,
// an array element, a parenthesized expression or condition, an index or a
// whole statement; all but the lowest operators as an assigned value; only
// atomic expressions elsewhere.
func level(n syntax.Node) slotLevel {
	if a, ok := n.Parent().(*syntax.Assign); ok && a.Value == n {
		return assignable
	}
	if nw, ok := n.Parent().(*syntax.New); ok && nw.Class == n {
		return className
	}
	if delimited(n) {
		return anyExpr
	}
	return operand
}

// needsParens reports whether expression x needs parentheses at a
// position of level l.
func needsParens(x syntax.Node, l slotLevel) bool {
	switch l {
	case operand:
		return loose(x)
	case className:
		switch x.Kind() {
		case syntax.KName, syntax.KVariable, syntax.KPropertyFetch, syntax.KStaticPropertyFetch, syntax.KArrayDimFetch:
			return false
		}
		return true // `new (expr)`: PHP 8.0 and later
	case assignable:
		b, ok := x.(*syntax.Binary)
		return ok && (b.Op.Kind == syntax.TAnd || b.Op.Kind == syntax.TOr || b.Op.Kind == syntax.TXor)
	}
	return false
}

func delimited(n syntax.Node) bool {
	switch p := n.Parent().(type) {
	case *syntax.Arg, *syntax.ArrayItem, *syntax.Paren, *syntax.ExprStmt,
		*syntax.Return, *syntax.Echo, *syntax.MatchArm:
		return true
	case *syntax.ArrayDimFetch:
		return p.Dim == n
	case *syntax.If:
		return p.Cond == n
	case *syntax.ElseIf:
		return p.Cond == n
	case *syntax.While:
		return p.Cond == n
	case *syntax.DoWhile:
		return p.Cond == n
	case *syntax.Switch:
		return p.Cond == n
	case *syntax.Match:
		return p.Cond == n
	}
	return false
}

// loose reports whether expression n binds looser than an operand.
func loose(n syntax.Node) bool {
	switch n.Kind() {
	case syntax.KVariable, syntax.KName, syntax.KConstFetch, syntax.KLiteral, syntax.KInterpolatedString,
		syntax.KMagicConst, syntax.KFuncCall, syntax.KMethodCall, syntax.KStaticCall, syntax.KPropertyFetch,
		syntax.KStaticPropertyFetch, syntax.KClassConstFetch, syntax.KArrayDimFetch, syntax.KArray,
		syntax.KList, syntax.KParen, syntax.KIsset, syntax.KEmpty:
		return false
	}
	return true
}

// render returns the fix replacing span, the match starting at matched,
// or false when the result would not be valid PHP or would drop a comment.
func (r *rule) render(m *matcher, matched syntax.Node, span syntax.Span) (diagnostic.Fix, bool) {
	ctx := m.ctx
	t := r.fix
	if t.snip == nil {
		return deletion(m, matched, span)
	}
	if _, inList := syntax.StmtListOf(matched.Parent()); t.snip.root == nil && !inList {
		return diagnostic.Fix{}, false // statements in a brace-less body
	}
	var b strings.Builder
	text := t.snip.text
	last := uint32(0)
	endsWithStmt := false
	for _, sl := range t.slots {
		start, end := sl.span.Start, sl.span.End
		bd := m.lookup(sl.name)
		repl := ctx.SpanText(bd.span)
		_, boundStmt := bd.node.(syntax.Stmt)
		switch {
		case boundStmt && !sl.isStmt:
			// A statement stands only where statements go.
			return diagnostic.Fix{}, false
		case boundStmt:
			start, end = sl.stmt.Start, sl.stmt.End
			endsWithStmt = end == uint32(len(text))
		case sl.spread && !bd.seq:
			repl = "..." + repl
		case sl.spread && len(bd.args) == 0:
			start, end, repl = sl.drop.Start, sl.drop.End, ""
		case bd.seq && !sl.spread:
			// A run of arguments stands only where arguments go.
			return diagnostic.Fix{}, false
		case needsParens(bd.node, sl.slot):
			repl = "(" + repl + ")"
		}
		b.WriteString(text[last:start])
		b.WriteString(repl)
		last = end
	}
	b.WriteString(text[last:])
	out := b.String()
	_, stmt := matched.(syntax.Stmt)
	switch {
	case stmt && t.snip.expr && !endsWithStmt:
		// An expression replacing a statement needs its `;`; a template
		// ending in a bound statement (`$S;`) carries its own.
		out += ";"
	case t.snip.expr && needsParens(t.snip.root, level(matched)):
		out = "(" + out + ")"
	}
	if !valid(out, t.snip.expr && !stmt, ctx.PHP) || astquery.DropsComment(ctx.File, span, out) {
		return diagnostic.Fix{}, false
	}
	return astquery.ReplaceFix(span, out), true
}

// deletion removes the matched statements (span, starting at matched), or
// the statement of the matched expression when it stands alone.
func deletion(m *matcher, matched syntax.Node, span syntax.Span) (diagnostic.Fix, bool) {
	if s, ok := matched.(syntax.Stmt); ok {
		return removal(m, s, span)
	}
	if s, ok := matched.Parent().(*syntax.ExprStmt); ok {
		return removal(m, s, s.Span())
	}
	return diagnostic.Fix{}, false
}

// removal removes statements covering stmts, the first of which is s.
func removal(m *matcher, s syntax.Stmt, stmts syntax.Span) (diagnostic.Fix, bool) {
	if _, ok := syntax.StmtListOf(s.Parent()); !ok {
		return diagnostic.Fix{}, false // a brace-less body would take the next statement
	}
	span := wholeLine(m.ctx.Src, stmts)
	if span == stmts {
		span = astquery.WithLeadingWhitespace(m.ctx.File, span)
	}
	if astquery.DropsComment(m.ctx.File, span, "") {
		return diagnostic.Fix{}, false
	}
	return astquery.ReplaceFix(span, ""), true
}

// valid reports whether out parses alone as one expression, or as
// statements.
func valid(out string, expr bool, php phpversion.Version) bool {
	src := snippetPrefix + out
	if expr {
		src += ";"
	}
	f := syntax.Parse("fix.php", []byte(src), syntax.Options{Version: php})
	n := len(slices.DeleteFunc(slices.Clone(f.Stmts), func(s syntax.Stmt) bool { return s.Kind() == syntax.KNop }))
	return len(f.Errors) == 0 && (n == 1 || !expr && n > 1)
}

// wholeLine extends span over its line, newline included, when nothing
// but blanks shares the line with it; else it returns span.
func wholeLine(src []byte, span syntax.Span) syntax.Span {
	start, end := span.Start, span.End
	for start > 0 && (src[start-1] == ' ' || src[start-1] == '\t') {
		start--
	}
	for int(end) < len(src) && (src[end] == ' ' || src[end] == '\t' || src[end] == '\r') {
		end++
	}
	if start > 0 && src[start-1] != '\n' || int(end) >= len(src) || src[end] != '\n' {
		return span
	}
	return syntax.Span{Start: start, End: end + 1}
}
