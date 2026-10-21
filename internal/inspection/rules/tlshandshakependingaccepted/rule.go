// Package tlshandshakependingaccepted implements TlsHandshakePendingAccepted.
package tlshandshakependingaccepted

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

const message = "Require a completed TLS handshake before writing."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "TlsHandshakePendingAccepted" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "stream_socket_enable_crypto") {
		return
	}
	b, ok := c.Parent().(*syntax.Binary)
	if !ok || b.Op.Kind != syntax.TIsNotIdentical || !semanticquery.IsFalse(b.Right) {
		return
	}
	branch, ok := b.Parent().(*syntax.If)
	if !ok {
		return
	}
	stream := semanticquery.CallArgument(c.Args, 0, "stream")
	var latest *syntax.FuncCall
	for _, event := range semanticquery.ScopeCallEvents(ctx, c) {
		p, ok := event.(*syntax.FuncCall)
		if ok && p.Span().End < c.Span().Start && semanticquery.NativeBuiltin(ctx, p, "stream_set_blocking") && semanticquery.SameStableLocal(ctx, semanticquery.CallArgument(p.Args, 0, "stream"), stream) {
			latest = p
		}
	}
	if latest == nil || !semanticquery.IsFalse(semanticquery.CallArgument(latest.Args, 1, "enable")) || !semanticquery.NativeCallSucceeded(ctx, latest, c) {
		return
	}
	stmts := []syntax.Stmt{branch.Body}
	if block, ok := branch.Body.(*syntax.Block); ok {
		stmts = block.Stmts
	}
	found := false
	for _, stmt := range stmts {
		syntax.Inspect(stmt, func(n syntax.Node) bool {
			if syntax.IsVariableScope(n) {
				return false
			}
			switch n.(type) {
			case *syntax.If, *syntax.While, *syntax.For, *syntax.Foreach, *syntax.DoWhile, *syntax.Switch, *syntax.Try:
				return false
			}
			if p, ok := n.(*syntax.FuncCall); ok && semanticquery.NativeBuiltin(ctx, p, "fwrite") && semanticquery.SameStableLocal(ctx, stream, semanticquery.CallArgument(p.Args, 0, "stream")) && !tlsHandshakeCompleted(ctx, branch.Body, stream, p) {
				found = true
			}
			return true
		})
	}
	if found {
		ctx.ReportNode(c, message)
	}
}

// tlsHandshakeCompleted reports whether a further enabling handshake call on
// the stream, written in body before write, is known to have returned true
// (neither false nor the pending 0) where write runs.
func tlsHandshakeCompleted(ctx *analysis.Context, body syntax.Stmt, stream syntax.Expr, write *syntax.FuncCall) bool {
	if semanticquery.NativeLabelled(ctx, write) {
		return false
	}
	return semanticquery.NativeEarlierCallProvesIn(ctx, body, write, func(x syntax.Expr) (types.Type, func(types.Type) bool) {
		c, ok := x.(*syntax.FuncCall)
		if !ok || !semanticquery.NativeBuiltin(ctx, c, "stream_socket_enable_crypto") || !semanticquery.SameStableLocal(ctx, stream, semanticquery.CallArgument(c.Args, 0, "stream")) {
			return types.Unknown, nil
		}
		if enabling, known := semanticquery.NativeTruth(ctx, semanticquery.CallArgument(c.Args, 1, "enable")); !known || !enabling {
			return types.Unknown, nil
		}
		// The handshake returns true, false or the pending 0; null stands for
		// 0: falsy and loosely equal to false, as 0 is.
		return types.Of("true", "false", "null"), func(t types.Type) bool { return t.OnlyOf("true") }
	})
}
