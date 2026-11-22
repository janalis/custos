// Package fiberstartedtwice implements the native FiberStartedTwice
// inspection.
package fiberstartedtwice

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/astquery"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
)

const message = "Start each fiber only once."

type rule struct{}

// New returns the FiberStartedTwice inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FiberStartedTwice" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.MethodCall)
	if fiberStartCallable(c) || semanticquery.FiberConstruction(ctx, c, "start") == nil {
		return
	}
	for _, first := range semanticquery.EarlierMethodCalls(ctx, c, c.Var, "Fiber", "start") {
		if !fiberStartCallable(first) && !fiberStartRebound(ctx, first, c) {
			ctx.ReportNode(c, message)
			return
		}
	}
}

// fiberStartCallable recognizes `$fiber->start(...)`, which only creates a
// closure and starts nothing.
func fiberStartCallable(c *syntax.MethodCall) bool {
	return c.Args.FirstClassCallable()
}

// fiberStartRebound reports whether the receiver variable of second may hold
// another fiber than at first: an assignment target (destructuring included),
// a foreach key/value or a catch variable binds it in between.
func fiberStartRebound(ctx *analysis.Context, first, second *syntax.MethodCall) bool {
	name := ""
	if v, ok := syntax.UnwrapParens(second.Var).(*syntax.Variable); ok {
		name = v.Name
	}
	scope := syntax.EnclosingVariableScope(second)
	rebound := false
	visit := func(n syntax.Node) bool {
		if rebound || (n != scope && syntax.IsVariableScope(n)) {
			return false
		}
		if n.Span().Start < first.Span().End || n.Span().Start >= second.Span().Start {
			return true
		}
		switch x := n.(type) {
		case *syntax.Assign:
			rebound = astquery.MentionsVariable(x.Var, name)
		case *syntax.Foreach:
			rebound = astquery.MentionsVariable(x.Key, name) || astquery.MentionsVariable(x.Value, name)
		case *syntax.Catch:
			rebound = x.Var != nil && x.Var.Name == name
		}
		return true
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return rebound
}
