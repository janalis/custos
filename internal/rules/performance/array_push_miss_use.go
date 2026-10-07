package performance

import (
	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// arrayPushMissUse reports single-element array_push() statements and
// `$a[count($a)] = …` appends.
type arrayPushMissUse struct{}

func init() { register(arrayPushMissUse{}) }

func (arrayPushMissUse) ID() string { return "ArrayPushMissUse" }

func (arrayPushMissUse) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall, syntax.KAssign}
}

func (r arrayPushMissUse) Check(ctx *analysis.Context, n syntax.Node) {
	switch n := n.(type) {
	case *syntax.FuncCall:
		r.checkPush(ctx, n)
	case *syntax.Assign:
		if ctx.Bool("REPORT_EXCESSIVE_COUNT_CALLS") {
			r.checkCount(ctx, n)
		}
	}
}

func (arrayPushMissUse) checkPush(ctx *analysis.Context, n *syntax.FuncCall) {
	call, _ := perfCall(ctx, n, "array_push") // D1
	if call == nil {
		return
	}
	if _, ok := call.Parent().(*syntax.ExprStmt); !ok { // D3
		return
	}
	// D2/D4; named arguments are skipped (spec Divergences).
	args, ok := perfPlainArgs(call.Args)
	if !ok || len(args) != 2 {
		return
	}
	// D8: `$a[] = $b` only matches array_push() on arrays (ArrayAccess
	// objects get offsetSet(null, …), strings and null behave differently).
	if !ctx.TypeOf(args[0]).IsArrayLike() {
		return
	}
	repl := ctx.Text(args[0]) + "[] = " + ctx.Text(args[1])
	span := call.Span()
	ctx.Report(span, "Use '"+repl+"' instead; it avoids a function call.", replaceFix(span, repl))
}

func (arrayPushMissUse) checkCount(ctx *analysis.Context, a *syntax.Assign) {
	if a.Op.Kind != syntax.TEqual { // D5
		return
	}
	dim, ok := a.Var.(*syntax.ArrayDimFetch)
	if !ok || dim.Dim == nil {
		return
	}
	call, _ := perfCall(ctx, dim.Dim, "count") // D6
	if call == nil {
		return
	}
	args, ok := perfPlainArgs(call.Args)
	if !ok || len(args) != 1 || !util.EquivalentFoldNames(ctx.File, dim.Var, args[0]) { // D7
		return
	}
	name := call.Name.(*syntax.Name)
	ctx.Report(util.NamePartSpan(name), "The index is redundant here; use '[]' to append.")
}
