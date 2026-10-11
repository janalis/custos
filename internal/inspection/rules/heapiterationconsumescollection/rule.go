// Package heapiterationconsumescollection implements the native HeapIterationConsumesCollection inspection.
package heapiterationconsumescollection

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Iterate a clone to preserve the heap."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "HeapIterationConsumesCollection" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KForeach} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	f := n.(*syntax.Foreach)
	if _, ok := syntax.UnwrapParens(f.Expr).(*syntax.Clone); ok {
		return
	}
	if iteratesLocalClone(ctx, f) {
		return
	}
	classes := ctx.Types().Native().TypeOf(f.Expr).Classes()
	if len(classes) != 1 {
		return
	}
	c := ctx.Index().Class(classes[0], ctx.PHP)
	if c == nil || !strings.HasPrefix(c.File, "stubs/") || (!ctx.Index().IsSubtype(c.FQN, "SplHeap", ctx.PHP) && !ctx.Index().IsSubtype(c.FQN, "SplPriorityQueue", ctx.PHP)) {
		return
	}
	ctx.ReportNode(f.Expr, message)
}

// iteratesLocalClone proves that the iterable is a plain local which can only
// hold a private copy: its single write in the scope is a `$copy = clone ...;`
// statement executed on every path to the loop, and every other mention is a
// read that can neither rebind nor alias the variable (another foreach over
// it, a method call on it, or count()). Anything else keeps the report.
func iteratesLocalClone(ctx *analysis.Context, f *syntax.Foreach) bool {
	local, ok := f.Expr.(*syntax.Variable)
	if !ok {
		return false
	}
	cloned, stable := false, true
	visit := func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.Include, *syntax.Eval:
			// Included and evaluated code shares the variable table.
			stable = false
		case *syntax.FuncCall:
			if name := ctx.GlobalFunctionName(n); name == "extract" || name == "parse_str" {
				stable = false
			}
		case *syntax.Variable:
			if n.NameExpr != nil || n.Name == "GLOBALS" {
				stable = false
			} else if n.Name == local.Name && !cloneLocalUse(ctx, n, f, &cloned) {
				stable = false
			}
		}
		return stable
	}
	if scope := syntax.EnclosingVariableScope(f); scope != nil {
		syntax.Inspect(scope, visit)
	} else {
		syntax.InspectFile(ctx.File, visit)
	}
	return cloned && stable
}

// cloneLocalUse accepts one mention of the tracked local. The defining clone
// assignment is accepted once and recorded through cloned.
func cloneLocalUse(ctx *analysis.Context, v *syntax.Variable, f *syntax.Foreach, cloned *bool) bool {
	switch p := v.Parent().(type) {
	case *syntax.Foreach:
		return p.Expr == v
	case *syntax.MethodCall:
		return p.Var == v
	case *syntax.Arg:
		list, _ := p.Parent().(*syntax.ArgList)
		call, ok := list.Parent().(*syntax.FuncCall)
		return ok && !p.Unpack && semanticquery.NativeBuiltin(ctx, call, "count")
	case *syntax.Assign:
		if p.Var != v || *cloned || p.ByRef || p.Op.Kind != syntax.TEqual {
			return false
		}
		if _, ok := syntax.UnwrapParens(p.Value).(*syntax.Clone); !ok {
			return false
		}
		if _, ok := p.Parent().(*syntax.ExprStmt); !ok {
			return false
		}
		*cloned = true
		return semanticquery.NativeDominates(p, f)
	}
	return false
}
