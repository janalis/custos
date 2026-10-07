package langmigration

import (
	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// getDebugTypeCanBeUsed reports `is_object($x) ? get_class($x) : gettype($x)`,
// which is get_debug_type($x) since PHP 8.0.
type getDebugTypeCanBeUsed struct{}

func init() { register(getDebugTypeCanBeUsed{}) }

func (getDebugTypeCanBeUsed) ID() string { return "GetDebugTypeCanBeUsed" }

func (getDebugTypeCanBeUsed) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KTernary} }

// gdtSingleArgCall returns the call and its only argument when e is directly
// a call named name (any case) with one argument resolving to the global function.
func gdtSingleArgCall(ctx *analysis.Context, e syntax.Expr, name string) (*syntax.FuncCall, syntax.Expr, bool) {
	call, ok := util.IsFuncNamedFold(e, name)
	if !ok {
		return nil, nil, false
	}
	args, ok := util.CallArgValues(call)
	if !ok || len(args) != 1 || !callsGlobalFunction(ctx, call, true) {
		return nil, nil, false
	}
	return call, args[0], true
}

func (getDebugTypeCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	t := n.(*syntax.Ternary)
	if t.Then == nil || ctx.PHP < phpver.PHP80 { // D1, D3
		return
	}
	cond, x, ok := gdtSingleArgCall(ctx, t.Cond, "is_object") // D2
	if !ok {
		return
	}
	if _, a, ok := gdtSingleArgCall(ctx, t.Then, "get_class"); !ok || !util.EquivalentFoldNames(ctx.File, a, x) { // D4
		return
	}
	if _, a, ok := gdtSingleArgCall(ctx, t.Else, "gettype"); !ok || !util.EquivalentFoldNames(ctx.File, a, x) { // D5
		return
	}
	qual, _, _ := util.CallName(cond)
	prefix := ""
	if qual == `\` {
		prefix = `\`
	}
	repl := prefix + "get_debug_type(" + ctx.Text(x) + ")"
	// F1: no fix. get_debug_type() names scalars and anonymous classes
	// differently from gettype()/get_class() (int/integer, float/double,
	// bool/boolean, null/NULL), so rewriting changes the produced string.
	ctx.Report(t.Span(), "Use '"+repl+"' instead (scalar type names differ from gettype()).")
}
