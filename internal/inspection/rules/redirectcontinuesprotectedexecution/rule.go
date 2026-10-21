// Package redirectcontinuesprotectedexecution implements the native RedirectContinuesProtectedExecution inspection.
package redirectcontinuesprotectedexecution

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

const message = "Terminate the unauthorized branch after redirecting."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "RedirectContinuesProtectedExecution" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	check(ctx, n, message)
}

func check(ctx *analysis.Context, n syntax.Node, message string) {
	c := n.(*syntax.FuncCall)

	report := false
	var fixes []diagnostic.Fix

	report = protectedRedirect(ctx, c)

	if report {
		ctx.ReportNode(c, message, fixes...)
	}
}

func protectedRedirect(ctx *analysis.Context, c *syntax.FuncCall) bool {
	if !semanticquery.CallsAnnotatedFunction(ctx, c, "@custos-protected") {
		return false
	}
	scope := syntax.EnclosingVariableScope(c)
	for p := c.Parent(); p != nil && p != scope; p = p.Parent() {
		if block, ok := p.(*syntax.Block); ok && redirectFallsThrough(ctx, scope, block.Stmts, c) {
			return true
		}
	}
	// Top-level statements have no enclosing block node.
	return scope == nil && redirectFallsThrough(ctx, scope, ctx.File.Stmts, c)
}

// redirectFallsThrough scans the statements preceding the protected
// call for a redirecting unauthorized branch that does not leave, whose
// predicate the conditions guarding the call do not prove truthy.
func redirectFallsThrough(ctx *analysis.Context, scope syntax.Node, stmts []syntax.Stmt, c *syntax.FuncCall) bool {
	for _, st := range stmts {
		branch, ok := st.(*syntax.If)
		if !ok || branch.Span().End > c.Span().Start || branch.Else != nil || len(branch.ElseIfs) > 0 || syntax.Terminates(branch.Body) {
			continue
		}
		unauthorized, ok := branch.Cond.(*syntax.Unary)
		if !ok || unauthorized.Op.Kind != syntax.TExclaim {
			continue
		}
		predicate, ok := unauthorized.Expr.(*syntax.Variable)
		if !ok || predicate.Name == "" || predicate.NameExpr != nil {
			continue
		}
		if redirectBranch(ctx, branch.Body) && !redirectPredicateRebound(ctx, scope, predicate.Name, branch.Span().Start, c.Span().Start) && !authorized(ctx, predicate.Name, c) {
			return true
		}
	}
	return false
}

// redirectBranch reports whether body emits a literal Location header
// in its own scope; nested function bodies run later, if at all.
func redirectBranch(ctx *analysis.Context, body syntax.Node) bool {
	found := false
	syntax.Inspect(body, func(child syntax.Node) bool {
		if found || syntax.IsVariableScope(child) {
			return false
		}
		header, ok := child.(*syntax.FuncCall)
		if !ok || semanticquery.NativeBuiltinName(ctx, header) != "header" {
			return true
		}
		value, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(header.Args, 0, "header"))
		if known && strings.HasPrefix(strings.ToLower(value), "location:") {
			found = true
		}
		return true
	})
	return found
}

// redirectPredicateRebound reports whether the authorization variable
// is written between the redirect branch and the protected call, which makes
// the relationship between the two unknown.
func redirectPredicateRebound(ctx *analysis.Context, scope syntax.Node, name string, from, to uint32) bool {
	for _, access := range flowquery.VarAccesses(ctx.File, scope, name) {
		if at := access.Var.Span().Start; access.Write && at >= from && at < to {
			return true
		}
	}
	return false
}

// authorized reports whether the conditions guarding c prove local
// name truthy there (or make c unreachable).
func authorized(ctx *analysis.Context, name string, c *syntax.FuncCall) bool {
	if semanticquery.NativeLabelled(ctx, c) {
		return false // a jump may land between a guard and the call
	}
	t := ctx.Types().Native().NarrowedVarAt(name, c, types.Bool)
	return t.IsUnknown() || t.OnlyOf("true")
}
