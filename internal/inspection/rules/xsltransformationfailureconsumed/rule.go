// Package xsltransformationfailureconsumed implements the native XslTransformationFailureConsumed inspection.
package xsltransformationfailureconsumed

import (
	"github.com/janalis/custos/internal/inspection/analysis"

	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
	"github.com/janalis/custos/internal/semantic/types"
)

const message = "Reject transformation failure before using the XML."

type rule struct{}

// New returns the XslTransformationFailureConsumed inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "XslTransformationFailureConsumed" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall, syntax.KEcho} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.File.Recovered() || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	var inputs []syntax.Expr
	switch c := n.(type) {
	case *syntax.FuncCall:
		if !semanticquery.NativeBuiltin(ctx, c, "file_put_contents") {
			return
		}
		inputs = []syntax.Expr{semanticquery.CallArgument(c.Args, 1, "data")}
	case *syntax.Echo:
		inputs = c.Exprs
	}
	for _, e := range inputs {
		c, ok := semanticquery.NativeLocalBuiltinResult(ctx, e, "false", "XSLTProcessor::transformToXML").(*syntax.MethodCall)
		if ok && semanticquery.NativeMethod(ctx, c, "XSLTProcessor", "transformToXML") && !semanticquery.NativeFailureExcludedSingle(ctx, e, types.Of("string", "null", "false"), "false") {
			ctx.ReportNode(n, message)
			return
		}
	}
}
