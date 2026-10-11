package arraypushmissuse

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// arrayPushMissUse reports single-element array_push() statements and
// `$a[count($a)] = …` appends.
type arrayPushMissUse struct{}

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
	call, _ := semanticquery.GlobalCall(ctx, n, "array_push") // D1
	if call == nil {
		return
	}
	if _, ok := call.Parent().(*syntax.ExprStmt); !ok { // D3
		return
	}
	// D2/D4; named arguments are skipped (spec Divergences).
	args, ok := astquery.ArgValues(call.Args)
	if !ok || len(args) != 2 {
		return
	}
	// D8: `$a[] = $b` only matches array_push() on arrays (ArrayAccess
	// objects get offsetSet(null, …), strings and null behave differently).
	if !ctx.TypeOf(args[0]).IsArrayLike() {
		return
	}
	value := ctx.Text(args[1])
	if apmBindsLooserThanAssign(args[1]) {
		value = "(" + value + ")"
	}
	repl := ctx.Text(args[0]) + "[] = " + value
	span := call.Span()
	ctx.Report(span, "Use '"+repl+"' instead; it avoids a function call.", astquery.ReplaceFix(span, repl))
}

// apmBindsLooserThanAssign reports whether e is an `and`/`or`/`xor`
// operation: the only expressions that would split apart as the right side
// of `=` (`$a[] = $b or $c` parses as `($a[] = $b) or $c`).
func apmBindsLooserThanAssign(e syntax.Expr) bool {
	b, ok := e.(*syntax.Binary)
	return ok && (b.Op.Kind == syntax.TAnd || b.Op.Kind == syntax.TOr || b.Op.Kind == syntax.TXor)
}

func (arrayPushMissUse) checkCount(ctx *analysis.Context, a *syntax.Assign) {
	if a.Op.Kind != syntax.TEqual { // D5
		return
	}
	dim, ok := a.Var.(*syntax.ArrayDimFetch)
	if !ok || dim.Dim == nil {
		return
	}
	call, _ := semanticquery.GlobalCall(ctx, dim.Dim, "count") // D6
	if call == nil {
		return
	}
	args, ok := astquery.ArgValues(call.Args)
	if !ok || len(args) != 1 || !astquery.EquivalentFoldNames(ctx.File, dim.Var, args[0]) { // D7
		return
	}
	if !apmFreshList(ctx, a, dim.Var) {
		return
	}
	name := call.Name.(*syntax.Name)
	ctx.Report(astquery.NamePartSpan(name), "The index is redundant here; use '[]' to append.")
}

// A newly assigned unkeyed literal has no sparse keys or hidden next-index
// history. Mere list-shaped types cannot establish that after unset().
func apmFreshList(ctx *analysis.Context, a *syntax.Assign, target syntax.Expr) bool {
	v, ok := target.(*syntax.Variable)
	stmt, standalone := a.Parent().(*syntax.ExprStmt)
	if !ok || v.NameExpr != nil || !standalone || flowquery.MayHaveSideEffects(a.Value) {
		return false
	}
	prev, _ := astquery.PrevStmt(ctx.File, stmt)
	es, ok := prev.(*syntax.ExprStmt)
	if !ok {
		return false
	}
	assign, ok := es.Expr.(*syntax.Assign)
	if !ok || assign.Op.Kind != syntax.TEqual || assign.ByRef || !astquery.Equivalent(ctx.File, assign.Var, target) {
		return false
	}
	arr, ok := syntax.UnwrapParens(assign.Value).(*syntax.Array)
	if !ok {
		return false
	}
	for _, item := range arr.Items {
		if item == nil || item.Key != nil || item.Unpack || item.ByRef {
			return false
		}
	}
	return true
}
