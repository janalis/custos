// Package arrayrandinvalidcount implements the native ArrayRandInvalidCount inspection.
package arrayrandinvalidcount

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Request a count within the array size."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ArrayRandInvalidCount" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "array_rand") {
		return
	}
	arg := semanticquery.CallArgument(c.Args, 1, "num")
	count, k := int64(1), true
	if arg != nil {
		count, k = semanticquery.NativeInt(ctx, arg)
		k = k && !rewritten(ctx, c, arg)
	}
	array := semanticquery.CallArgument(c.Args, 0, "array")
	entries, known := semanticquery.NativeArrayEntries(ctx, array)
	known = known && !rewritten(ctx, c, array)
	if k && (count < 1 || known && count > int64(len(entries))) {
		ctx.ReportNode(c, message)
	}
}

// rewritten reports whether e is a plain variable that its scope changes in
// a way the tracked value does not follow: `++`/`--` on it or on one of its
// offsets, a foreach that writes into it, a by-reference capture or array
// slot, or extract(). A variable whose value an assignment copies into e,
// directly or through further copies, counts too. The scan ignores statement
// order.
func rewritten(ctx *analysis.Context, call *syntax.FuncCall, e syntax.Expr) bool {
	v, ok := syntax.UnwrapParens(e).(*syntax.Variable)
	if !ok || v.NameExpr != nil {
		return false
	}
	hidden := map[string]bool{}    // names written outside plain assignments
	feeds := map[string][]string{} // assigned name -> variables read by the values
	extract := false
	visit := func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.IncDec:
			mark(hidden, n.Var)
		case *syntax.Foreach:
			mark(hidden, n.Key)
			mark(hidden, n.Value)
		case *syntax.ClosureUse:
			if n.ByRef {
				hidden[n.Var.Name] = true
			}
		case *syntax.ArrayItem:
			if n.ByRef {
				mark(hidden, n.Value)
			}
		case *syntax.FuncCall:
			_, is := astquery.IsFuncNamedFold(n, "extract")
			extract = extract || is
		case *syntax.Assign:
			if target, ok := syntax.UnwrapParens(n.Var).(*syntax.Variable); ok {
				read := map[string]bool{}
				mark(read, n.Value)
				for name := range read {
					feeds[target.Name] = append(feeds[target.Name], name)
				}
			}
		}
		return !extract
	}
	if scope := syntax.EnclosingVariableScope(call); scope != nil {
		syntax.Inspect(scope, visit)
	} else {
		syntax.InspectFile(ctx.File, visit)
	}
	seen := map[string]bool{v.Name: true}
	for queue := []string{v.Name}; len(queue) > 0 && !extract; queue = queue[1:] {
		if hidden[queue[0]] {
			return true
		}
		for _, name := range feeds[queue[0]] {
			if !seen[name] {
				seen[name] = true
				queue = append(queue, name)
			}
		}
	}
	return extract
}

// mark records every variable named under n; n may be an absent foreach key.
func mark(names map[string]bool, n syntax.Expr) {
	if n == nil {
		return
	}
	syntax.Inspect(n, func(x syntax.Node) bool {
		if v, ok := x.(*syntax.Variable); ok {
			names[v.Name] = true
		}
		return true
	})
}
