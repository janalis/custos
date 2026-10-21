// Package nonblockingsocketconnectpendingrejected implements NonblockingSocketConnectPendingRejected.
package nonblockingsocketconnectpendingrejected

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Handle a pending nonblocking connection."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "NonblockingSocketConnectPendingRejected" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "socket_connect") {
		return
	}
	branch := semanticquery.FailureBranch(c)
	socket := semanticquery.CallArgument(c.Args, 0, "socket")
	if branch == nil || nonblockingConnectInspected(ctx, branch.Body, socket) {
		return
	}
	var latest *syntax.FuncCall
	for _, event := range semanticquery.ScopeCallEvents(ctx, c) {
		p, ok := event.(*syntax.FuncCall)
		if !ok || p.Span().End >= c.Span().Start || !semanticquery.SameStableLocal(ctx, semanticquery.CallArgument(p.Args, 0, "socket"), socket) {
			continue
		}
		if semanticquery.NativeBuiltin(ctx, p, "socket_set_block") || semanticquery.NativeBuiltin(ctx, p, "socket_set_nonblock") {
			latest = p
		}
	}
	// The last mode switch must be the nonblocking one and must have run:
	// either its success guards the connect or it is a dominating statement.
	if latest != nil && semanticquery.NativeBuiltin(ctx, latest, "socket_set_nonblock") && (semanticquery.NativeCallSucceeded(ctx, latest, c) || semanticquery.NativeDominates(latest, c)) {
		ctx.ReportNode(c, message)
	}
}

// nonblockingConnectInspected reports whether the failure branch may look at
// the reason of the failure: a direct error query, or a call that is not a
// builtin function and can reach the socket through its receiver or arguments.
func nonblockingConnectInspected(ctx *analysis.Context, body syntax.Node, socket syntax.Expr) bool {
	var names []string
	syntax.Inspect(socket, func(n syntax.Node) bool {
		if v, ok := n.(*syntax.Variable); ok && v.NameExpr == nil {
			names = append(names, v.Name)
		}
		return true
	})
	reaches := func(parts ...syntax.Node) bool {
		for _, part := range parts {
			for _, name := range names {
				if astquery.MentionsVariable(part, name) {
					return true
				}
			}
		}
		return false
	}
	inspected := false
	syntax.Inspect(body, func(n syntax.Node) bool {
		if syntax.IsVariableScope(n) {
			return false
		}
		switch x := n.(type) {
		case *syntax.FuncCall:
			switch semanticquery.NativeBuiltinName(ctx, x) {
			case "socket_last_error", "socket_get_option", "socket_getopt":
				inspected = true
			case "":
				inspected = inspected || reaches(x.Name, x.Args)
			}
		case *syntax.MethodCall:
			inspected = inspected || reaches(x.Var, x.Args)
		case *syntax.StaticCall:
			inspected = inspected || reaches(x.Class, x.Args)
		}
		return !inspected
	})
	return inspected
}
