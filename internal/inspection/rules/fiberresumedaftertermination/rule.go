// Package fiberresumedaftertermination implements FiberResumedAfterTermination.
package fiberresumedaftertermination

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Resume a fiber before it terminates."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FiberResumedAfterTermination" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	check(ctx, n, message)
}

func check(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.MethodCall)
	construct := semanticquery.FiberConstruction(ctx, c, "resume")
	receiver, local := syntax.UnwrapParens(c.Var).(*syntax.Variable)
	if construct == nil || !local || fiberResumeCallableOnly(c) || !fiberResumeStarted(ctx, c) || fiberResumeRebound(ctx, c, receiver.Name, construct) {
		return
	}
	params, body := semanticquery.NativeCallback(ctx, semanticquery.CallArgument(construct.Args, 0, "callback"))
	if len(params) > 0 {
		return
	}
	var value syntax.Expr
	if e, ok := body.(syntax.Expr); ok {
		value = e
	} else if b, ok := body.(*syntax.Block); ok && len(b.Stmts) == 1 {
		if r, ok := b.Stmts[0].(*syntax.Return); ok {
			value = r.Expr
		}
	}
	if semanticquery.IsLiteral(value) {
		ctx.ReportNode(c, message)
	}
}

// fiberResumeStarted requires an executed start() before the resume.
func fiberResumeStarted(ctx *analysis.Context, resume *syntax.MethodCall) bool {
	for _, start := range semanticquery.EarlierMethodCalls(ctx, resume, resume.Var, "Fiber", "start") {
		if !fiberResumeCallableOnly(start) {
			return true
		}
	}
	return false
}

// fiberResumeCallableOnly recognizes `$fiber->method(...)`, which creates a
// closure without invoking the method.
func fiberResumeCallableOnly(c *syntax.MethodCall) bool {
	for _, arg := range c.Args.Args {
		if _, ok := arg.(*syntax.VariadicPlaceholder); ok {
			return true
		}
	}
	return false
}

// fiberResumeRebound reports whether the receiver may hold another fiber than
// construct. Between the allocation and the resume, a foreach target, a
// destructuring target, a catch or static variable or unset() gives the
// variable a new value; a reference to it (reference assignment, by-reference
// closure capture, `global`) anywhere before the resume lets other code do so.
func fiberResumeRebound(ctx *analysis.Context, resume *syntax.MethodCall, name string, construct *syntax.New) bool {
	from, to := construct.Span().End, resume.Span().Start
	mentions := func(target syntax.Node, anywhere bool) bool {
		found := false
		syntax.Inspect(target, func(n syntax.Node) bool {
			if v, ok := n.(*syntax.Variable); ok && v.Name == name && v.Span().Start < to && (anywhere || v.Span().Start >= from) {
				found = true
			}
			return !found
		})
		return found
	}
	scope := syntax.EnclosingVariableScope(resume)
	rebound := false
	visit := func(n syntax.Node) bool {
		if rebound || n.Span().Start >= to {
			return false
		}
		switch x := n.(type) {
		case *syntax.Assign:
			if x.ByRef {
				rebound = mentions(x, true)
				break
			}
			switch x.Var.(type) {
			case *syntax.Array, *syntax.List:
				rebound = mentions(x.Var, false)
			}
		case *syntax.Foreach:
			rebound = mentions(x.Value, false) || (x.Key != nil && mentions(x.Key, false))
		case *syntax.Catch:
			rebound = x.Var != nil && mentions(x.Var, false)
		case *syntax.StaticVar:
			rebound = mentions(x, false)
		case *syntax.Unset:
			rebound = mentions(x, false)
		case *syntax.Global:
			rebound = mentions(x, true)
		case *syntax.Closure:
			for _, use := range x.Uses {
				rebound = rebound || (use.ByRef && mentions(use, true))
			}
		}
		return !rebound && (n == scope || !syntax.IsVariableScope(n))
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return rebound
}
