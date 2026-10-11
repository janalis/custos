package elvisoperatorcanbeused

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/php/syntax"
)

// elvisOperatorCanBeUsed reports `$a ? $a : $b`, better written `$a ?: $b`.
type elvisOperatorCanBeUsed struct{}

func (elvisOperatorCanBeUsed) ID() string               { return "ElvisOperatorCanBeUsed" }
func (elvisOperatorCanBeUsed) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KTernary} }

func (elvisOperatorCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	t := n.(*syntax.Ternary)
	if t.Cond == nil || t.Then == nil || t.Else == nil { // D1, D2
		return
	}
	if t.Else.Span().Len() == 0 { // error recovery: missing else branch
		return
	}
	span := t.Span()
	if !astquery.EquivalentFoldNames(ctx.File, syntax.UnwrapParens(t.Cond), syntax.UnwrapParens(t.Then)) { // D3
		return
	}
	if constNames(ctx, t.Cond) != constNames(ctx, t.Then) { // D3: constant names are case-sensitive
		return
	}
	// E4: the short ternary evaluates the condition once; with calls or
	// other side effects that drops the second evaluation.
	if flowquery.MayHaveSideEffects(t.Cond) {
		return
	}
	repl := ctx.Text(t.Cond) + " ?: " + ctx.Text(t.Else) // D4
	ctx.Report(span, "Use the short ternary: '"+repl+"'.", diagnostic.Fix{
		Title: "Replace with '" + repl + "'",
		Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: span, NewText: repl}} },
	})
}

// constNames joins, in source order, the class-constant names fetched under
// n. The shared equivalence folds every keyword token, and a constant named
// like a keyword (`Foo::Array`) is lexed as one, so the names are compared
// exactly here; only the `class` pseudo-constant is case-insensitive.
func constNames(ctx *analysis.Context, n syntax.Node) string {
	var sb strings.Builder
	syntax.Inspect(n, func(c syntax.Node) bool {
		if f, ok := c.(*syntax.ClassConstFetch); ok {
			if id, ok := f.Name.(*syntax.Identifier); ok {
				name := ctx.Text(id)
				if strings.EqualFold(name, "class") {
					name = "class"
				}
				sb.WriteString(name)
				sb.WriteByte(':')
			}
		}
		return true
	})
	return sb.String()
}
