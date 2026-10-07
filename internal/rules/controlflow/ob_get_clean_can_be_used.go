package controlflow

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// obGetCleanCanBeUsed reports `ob_get_contents()` immediately followed by an
// `ob_end_clean();` statement.
type obGetCleanCanBeUsed struct{}

func init() { register(obGetCleanCanBeUsed{}) }

func (obGetCleanCanBeUsed) ID() string { return "ObGetCleanCanBeUsed" }

func (obGetCleanCanBeUsed) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func (obGetCleanCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	end, ok := util.IsFuncNamedFold(n, "ob_end_clean") // D1
	if !ok {
		return
	}
	stmt, ok := end.Parent().(*syntax.ExprStmt)
	if !ok || stmt.Expr != syntax.Expr(end) {
		return
	}
	prev, ok := util.PrevStmtNoDoc(ctx.File, stmt) // D2
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
			if c, ok := util.IsFuncNamedFold(x, "ob_get_contents"); ok {
				if get == nil {
					get = c
				}
				count++
			}
		}
		return true
	})
	if get == nil || count > 1 || !obIsBuiltin(ctx, end, "ob_end_clean") || !obIsBuiltin(ctx, get, "ob_get_contents") { // D4
		return
	}
	name := get.Name.(*syntax.Name)
	ctx.ReportNode(get, "Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().", analysis.Fix{
		Title: "Use ob_get_clean()",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{
				{Span: name.Span(), NewText: util.QualifiedBuiltinFor(ctx, "ob_get_clean", get)},
				{Span: stmt.Span(), NewText: ""},
			}
		},
	})
}

// obIsBuiltin reports whether call resolves to the global function global.
func obIsBuiltin(ctx *analysis.Context, call *syntax.FuncCall, global string) bool {
	f := ctx.Types().ResolveFunction(call)
	return f != nil && strings.EqualFold(strings.TrimPrefix(f.FQN, `\`), global)
}
