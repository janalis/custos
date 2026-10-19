// Package intdivzerodivisor implements the native IntdivZeroDivisor inspection.
package intdivzerodivisor

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply a nonzero integer divisor."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "IntdivZeroDivisor" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "intdiv") {
		return
	}
	divisor := semanticquery.CallArgument(c.Args, 1, "num2")
	v, k := semanticquery.NativeInt(ctx, divisor)
	if k && v == 0 && !rewrittenBeforeCall(ctx.File, c, divisor) {
		ctx.ReportNode(c, message)
	}
}

// rewrittenBeforeCall reports whether a variable divisor may have been changed
// before the call runs by a write the value resolution does not model: `++`
// or `--`, a by-reference closure capture, a destructuring foreach target, a
// catch binding or extract(). A counter that starts at zero would otherwise
// still look like a proven zero.
func rewrittenBeforeCall(f *syntax.File, call *syntax.FuncCall, divisor syntax.Expr) bool {
	v, ok := syntax.UnwrapParens(divisor).(*syntax.Variable)
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
			target, plain := w.Var.(*syntax.Variable)
			writes = plain && target.Name == v.Name
		case *syntax.Closure:
			for _, use := range w.Uses {
				writes = writes || use.ByRef && use.Var.Name == v.Name
			}
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
		// A nested function owns its variables; only a by-reference capture
		// shares one with the scope of the call.
		return !rewritten && (x == scope || !syntax.IsVariableScope(x))
	}
	if scope == nil {
		syntax.InspectFile(f, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return rewritten
}
