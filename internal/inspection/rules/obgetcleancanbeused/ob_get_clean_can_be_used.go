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
func (obGetCleanCanBeUsed) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (obGetCleanCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	end, ok := astquery.IsFuncNamedFold(n, "ob_end_clean") // D1
	if !ok {
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
			if c, ok := astquery.IsFuncNamedFold(x, "ob_get_contents"); ok {
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
	name := get.Name.(*syntax.Name)
	ctx.ReportNode(get, "Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().", diagnostic.Fix{
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
