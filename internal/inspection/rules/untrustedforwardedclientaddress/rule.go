// Package untrustedforwardedclientaddress implements the native UntrustedForwardedClientAddress inspection.
package untrustedforwardedclientaddress

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Validate the trusted proxy boundary before trusting forwarded addresses."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "UntrustedForwardedClientAddress" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	check(ctx, n, message)
}

func check(ctx *analysis.Context, n syntax.Node, message string) {
	binary := n.(*syntax.Binary)
	if binary.Op.Kind != syntax.TIsIdentical && binary.Op.Kind != syntax.TIsEqual {
		return
	}
	for _, operand := range []syntax.Expr{binary.Left, binary.Right} {
		fetch, ok := operand.(*syntax.ArrayDimFetch)
		if !ok {
			continue
		}
		if forwardedServerKey(ctx, fetch) == "HTTP_X_FORWARDED_FOR" {
			for p := binary.Parent(); p != nil; p = p.Parent() {
				if branch, ok := p.(*syntax.If); ok && branch.Cond != nil && syntax.UnwrapParens(branch.Cond) == binary && protectedBranch(ctx, branch.Body) {
					if !forwardedProxyGuarded(ctx, branch) {
						ctx.ReportNode(binary, message)
					}
					return
				}
			}
		}
	}
}

func protectedBranch(ctx *analysis.Context, body syntax.Stmt) bool {
	found := false
	syntax.Inspect(body, func(node syntax.Node) bool {
		if node != body && syntax.IsVariableScope(node) {
			return false
		}
		if call, ok := node.(*syntax.FuncCall); ok && semanticquery.CallsAnnotatedFunction(ctx, call, "@custos-protected") {
			found = true
		}
		return true
	})
	return found
}

// forwardedProxyGuarded reports whether a condition that may validate the
// proxy boundary already decides whether branch runs: the condition of an
// enclosing if chain, or of an earlier if in an enclosing statement list one
// of whose arms leaves that list.
func forwardedProxyGuarded(ctx *analysis.Context, branch *syntax.If) bool {
	for n := syntax.Node(branch); n != nil && !syntax.IsVariableScope(n); n = n.Parent() {
		if outer, ok := n.(*syntax.If); ok && outer != branch && forwardedProxyChain(ctx, outer) {
			return true
		}
		if st, ok := n.(syntax.Stmt); ok {
			for prev, exists := astquery.PrevStmt(ctx.File, st); exists; prev, exists = astquery.PrevStmt(ctx.File, prev) {
				if guard, ok := prev.(*syntax.If); ok && forwardedChainLeaves(guard) && forwardedProxyChain(ctx, guard) {
					return true
				}
			}
		}
	}
	return false
}

// forwardedProxyChain reports whether any condition of the if chain may
// validate the proxy boundary.
func forwardedProxyChain(ctx *analysis.Context, chain *syntax.If) bool {
	if forwardedProxyTest(ctx, chain.Cond) {
		return true
	}
	for _, arm := range chain.ElseIfs {
		if forwardedProxyTest(ctx, arm.Cond) {
			return true
		}
	}
	return false
}

// forwardedChainLeaves reports whether some arm of the if chain never falls
// through to the statement after it.
func forwardedChainLeaves(chain *syntax.If) bool {
	if syntax.Terminates(chain.Body) || (chain.Else != nil && syntax.Terminates(chain.Else.Body)) {
		return true
	}
	for _, arm := range chain.ElseIfs {
		if syntax.Terminates(arm.Body) {
			return true
		}
	}
	return false
}

// forwardedProxyTest reports whether cond reads the connecting peer address
// or calls anything but a builtin function: either may be a trusted-proxy
// check this rule does not model.
func forwardedProxyTest(ctx *analysis.Context, cond syntax.Expr) bool {
	found := false
	syntax.Inspect(cond, func(node syntax.Node) bool {
		switch x := node.(type) {
		case *syntax.ArrayDimFetch:
			found = found || forwardedServerKey(ctx, x) == "REMOTE_ADDR"
		case *syntax.FuncCall:
			function := ctx.Types().ResolveFunction(x)
			found = found || function == nil || !function.Builtin
		case *syntax.MethodCall, *syntax.StaticCall:
			found = true
		}
		return !found
	})
	return found
}

// forwardedServerKey returns the constant key of a $_SERVER element read, or
// "" for any other expression.
func forwardedServerKey(ctx *analysis.Context, fetch *syntax.ArrayDimFetch) string {
	server, ok := fetch.Var.(*syntax.Variable)
	key, known := semanticquery.NativeString(ctx, fetch.Dim)
	if !ok || !known || server.Name != "_SERVER" {
		return ""
	}
	return key
}
