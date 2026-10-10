// Package sysvmessageserializationmismatch implements the native SysvMessageSerializationMismatch inspection.
package sysvmessageserializationmismatch

import (
	"strings"

	"custos/internal/inspection/analysis"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use matching serialization flags for the message."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SysvMessageSerializationMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "msg_receive") {
		return
	}
	send := semanticquery.ExpansionDMessage(ctx, c)
	if send == nil {
		return
	}
	a := semanticquery.CallArgument(send.Args, 3, "serialize")
	b := semanticquery.CallArgument(c.Args, 5, "unserialize")
	x, xk := true, true
	if a != nil {
		x, xk = semanticquery.NativeTruth(ctx, a)
	}
	y, yk := true, true
	if b != nil {
		y, yk = semanticquery.NativeTruth(ctx, b)
	}
	// Receiving serialized bytes without decoding is a valid raw transport.
	if xk && yk && !x && y {
		text, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(send.Args, 2, "message"))
		// PHP serialized values have either N; or a type/length separator.
		// The full serialization grammar and unknown payloads remain unproven.
		if !known || text == "N;" || strings.Contains(text, ":") {
			return
		}
		ctx.ReportNode(n, message)
	}
}
