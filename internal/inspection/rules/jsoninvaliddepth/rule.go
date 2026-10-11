// Package jsoninvaliddepth implements the native JsonInvalidDepth inspection.
package jsoninvaliddepth

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply a supported JSON depth."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "JsonInvalidDepth" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)
	if name != "json_decode" && name != "json_encode" {
		return
	}
	depth := semanticquery.CallArgument(c.Args, 2, "depth")
	d, k := semanticquery.NativeInt(ctx, depth)
	if k && (d <= 0 || name == "json_decode" && d > 2147483647) && !rewrittenBeforeCall(ctx.File, c, depth) {
		ctx.ReportNode(c, message)
	}
}

// rewrittenBeforeCall reports whether a variable depth may have been changed
// before the call runs by a write the value resolution does not model: `++`
// or `--`, a by-reference closure capture, a destructuring foreach target, a
// catch binding or extract(). A depth counter that starts at zero would
// otherwise still look like a proven zero.
func rewrittenBeforeCall(f *syntax.File, call *syntax.FuncCall, depth syntax.Expr) bool {
	v, ok := syntax.UnwrapParens(depth).(*syntax.Variable)
	if !ok {
		return false
	}
	// A write placed after the call still precedes it on the next iteration
	// of any loop enclosing the call.
	looped := false
	var scope syntax.Node
	for p := call.Parent(); p != nil && scope == nil; p = p.Parent() {
		switch p.(type) {
		case *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach:
			looped = true
		case *syntax.Function, *syntax.Method, *syntax.Closure, *syntax.ArrowFunction:
			scope = p
		}
	}
	rewritten := false
	visit := func(x syntax.Node) bool {
		writes := false
		switch w := x.(type) {
		case *syntax.IncDec:
			target, plain := syntax.UnwrapParens(w.Var).(*syntax.Variable)
			writes = plain && target.Name == v.Name
		case *syntax.ClosureUse:
			writes = w.ByRef && w.Var.Name == v.Name
		case *syntax.Foreach:
			writes = astquery.MentionsVariable(w.Key, v.Name) || astquery.MentionsVariable(w.Value, v.Name)
		case *syntax.Catch:
			writes = w.Var != nil && w.Var.Name == v.Name
		case *syntax.FuncCall:
			_, writes = astquery.IsFuncNamedFold(w, "extract")
		}
		if writes && (looped || x.Span().Start < call.Span().Start) {
			rewritten = true
		}
		return !rewritten
	}
	if scope == nil {
		syntax.InspectFile(f, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return rewritten
}
