package iscountablecanbeused

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// isCountableCanBeUsed reports `is_array($v) || $v instanceof Countable`,
// which is is_countable($v).
type isCountableCanBeUsed struct{}

func (isCountableCanBeUsed) ID() string               { return "IsCountableCanBeUsed" }
func (isCountableCanBeUsed) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func isBooleanOr(n syntax.Node) (*syntax.Binary, bool) {
	b, ok := n.(*syntax.Binary)
	return b, ok && b.Op.Kind == syntax.TBooleanOr
}

func (isCountableCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP74 { // E1
		return
	}
	call := n.(*syntax.FuncCall)
	if ctx.GlobalFunctionName(call) != "is_array" { // D1 (any case, global function)
		return
	}
	args, ok := astquery.CallArgValues(call)
	if !ok || len(args) != 1 { // D2
		return
	}
	top, ok := isBooleanOr(call.Parent()) // D3
	if !ok {
		return
	}
	for { // D4
		p, _ := astquery.ParentSkipParens(top)
		b, ok := isBooleanOr(p)
		if !ok {
			break
		}
		top = b
	}
	var found *syntax.Instanceof
	var visit func(e syntax.Expr) // D5, D6
	visit = func(e syntax.Expr) {
		if found != nil {
			return
		}
		e = syntax.UnwrapParens(e)
		if b, ok := isBooleanOr(e); ok {
			visit(b.Left)
			visit(b.Right)
			return
		}
		io, ok := e.(*syntax.Instanceof)
		if !ok {
			return
		}
		if !semanticquery.NamesGlobalClass(ctx, io.Class, "Countable") {
			return
		}
		if astquery.EquivalentFoldNames(ctx.File, io.Expr, args[0]) {
			found = io
		}
	}
	visit(top)
	if found == nil {
		return
	}
	ctx.ReportNode(call, "Use 'is_countable("+ctx.Text(found.Expr)+")' instead of the is_array()/instanceof Countable pair.") // D7
}
