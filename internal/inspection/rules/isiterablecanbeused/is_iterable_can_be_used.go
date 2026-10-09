package isiterablecanbeused

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// isIterableCanBeUsed reports `is_array($v) || $v instanceof Traversable`
// pairs that is_iterable() covers.
type isIterableCanBeUsed struct{}

func (isIterableCanBeUsed) ID() string { return "IsIterableCanBeUsed" }
func (isIterableCanBeUsed) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func (isIterableCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP71 { // E1
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
	argVal := args[0]
	parent, ok := call.Parent().(*syntax.Binary)
	if !ok || parent.Op.Kind != syntax.TBooleanOr { // D3
		return
	}
	top := parent // D4
	for {
		p, _ := astquery.ParentSkipParens(top)
		b, ok := p.(*syntax.Binary)
		if !ok || b.Op.Kind != syntax.TBooleanOr {
			break
		}
		top = b
	}
	var subject syntax.Expr
	var walk func(b *syntax.Binary) bool // D5, D6
	walk = func(b *syntax.Binary) bool {
		for _, op := range [2]syntax.Expr{b.Left, b.Right} {
			e := syntax.UnwrapParens(op)
			if inner, ok := e.(*syntax.Binary); ok && inner.Op.Kind == syntax.TBooleanOr {
				if walk(inner) {
					return true
				}
				continue
			}
			io, ok := e.(*syntax.Instanceof)
			if !ok || !semanticquery.NamesGlobalClass(ctx, io.Class, "Traversable") || !astquery.EquivalentFoldNames(ctx.File, io.Expr, argVal) {
				continue
			}
			subject = io.Expr
			return true
		}
		return false
	}
	if walk(top) { // D7
		ctx.ReportNode(call, "Use 'is_iterable("+ctx.Text(subject)+")' instead of the is_array()/instanceof Traversable pair.")
	}
}
