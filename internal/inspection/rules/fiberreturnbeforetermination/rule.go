// Package fiberreturnbeforetermination implements FiberReturnBeforeTermination.
package fiberreturnbeforetermination

import (
	"slices"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Read the fiber return after termination."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FiberReturnBeforeTermination" }
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
	construct := semanticquery.FiberConstruction(ctx, c, "getReturn")
	if construct == nil || !fiberStarted(ctx, c) {
		return
	}
	params, body := semanticquery.NativeCallback(ctx, semanticquery.CallArgument(construct.Args, 0, "callback"))
	if len(params) > 0 {
		return
	}
	if !suspends(ctx, body) {
		return
	}
	if fiberUntouched(ctx, c, construct) {
		ctx.ReportNode(c, message)
	}
}

// fiberUntouched proves that nothing but the allocation and
// dominating start calls names the receiver before the read at c executes.
// Any other mention (a resume in a loop or branch, a state guard, an alias, an
// argument, a capture) may complete the fiber, so it abandons the proof.
func fiberUntouched(ctx *analysis.Context, c *syntax.MethodCall, construct *syntax.New) bool {
	// The allocation only resolves for a plain variable receiver.
	read := syntax.UnwrapParens(c.Var)
	receiver, _ := astquery.PlainVariableName(read)
	allowed := []syntax.Node{read}
	for _, start := range semanticquery.EarlierMethodCalls(ctx, c, c.Var, "Fiber", "start") {
		allowed = append(allowed, syntax.UnwrapParens(start.Var))
	}
	scope := syntax.EnclosingVariableScope(c)
	// Code after the read cannot run first, unless a loop repeats it.
	limit := c.Span().End
	for p := c.Parent(); p != nil && p != scope; p = p.Parent() {
		switch p.(type) {
		case *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach:
			limit = p.Span().End
		}
	}
	untouched := true
	visit := func(n syntax.Node) bool {
		if !untouched || n.Span().Start >= limit {
			return false
		}
		switch x := n.(type) {
		case *syntax.Function, *syntax.Method, *syntax.PropertyHook, *syntax.Closure:
			// Own variables; a captured fiber already fails the alias proof.
			return false
		case *syntax.FuncCall:
			// These expose every local by name without mentioning it.
			if name, ok := x.Name.(*syntax.Name); ok && slices.Contains([]string{"compact", "extract", "get_defined_vars"}, strings.ToLower(strings.TrimPrefix(name.Value, `\`))) {
				untouched = false
			}
		case *syntax.Variable:
			if x.NameExpr != nil {
				untouched = false
			} else if x.Name == receiver && !slices.Contains(allowed, syntax.Node(x)) {
				assign, ok := x.Parent().(*syntax.Assign)
				if !ok || assign.Var != syntax.Expr(x) || syntax.UnwrapParens(assign.Value) != syntax.Expr(construct) {
					untouched = false
				}
			}
		}
		return untouched
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Children(scope, func(child syntax.Node) { syntax.Inspect(child, visit) })
	}
	return untouched
}

func fiberStarted(ctx *analysis.Context, c *syntax.MethodCall) bool {
	return len(semanticquery.EarlierMethodCalls(ctx, c, c.Var, "Fiber", "start")) > 0
}

func suspends(ctx *analysis.Context, body syntax.Node) bool {
	if e, ok := body.(syntax.Expr); ok {
		return semanticquery.ExpansionCStatic(ctx, syntax.UnwrapParens(e), "Fiber", "suspend")
	}
	block, ok := body.(*syntax.Block)
	if !ok || len(block.Stmts) == 0 {
		return false
	}
	stmt, ok := block.Stmts[0].(*syntax.ExprStmt)
	return ok && semanticquery.ExpansionCStatic(ctx, syntax.UnwrapParens(stmt.Expr), "Fiber", "suspend")
}
