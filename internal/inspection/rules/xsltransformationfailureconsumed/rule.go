// Package xsltransformationfailureconsumed implements the native XslTransformationFailureConsumed inspection.
package xsltransformationfailureconsumed

import (
	"custos/internal/inspection/analysis"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

const message = "Reject transformation failure before using the XML."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "XslTransformationFailureConsumed" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall, syntax.KEcho} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
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
		if ok && semanticquery.NativeMethod(ctx, c, "XSLTProcessor", "transformToXML") && !ctx.Flow().Excludes(e, "false") && (semanticquery.NativeFailureFlowOnly(ctx, e) || !semanticquery.NativeNarrowedExcludes(ctx, e, types.Of("string", "null", "false"), "false")) {
			ctx.ReportNode(n, message)
			return
		}
	}
}
