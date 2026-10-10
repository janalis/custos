// Package soaplastrequestwithouttracing implements the native SoapLastRequestWithoutTracing inspection.
package soaplastrequestwithouttracing

import (
	"custos/internal/inspection/analysis"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Enable SOAP tracing before retrieving the request."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SoapLastRequestWithoutTracing" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.MethodCall)
	if !semanticquery.ExpansionDUnaliased(ctx, c.Var, n) {
		return
	}
	if !semanticquery.NativeMethod(ctx, c, "SoapClient", "__getLastRequest") {
		return
	}
	creation := semanticquery.NativeConstruction(ctx, c.Var, "SoapClient")
	if creation == nil {
		return
	}
	a := semanticquery.CallArgument(creation.Args, 1, "options")
	if a == nil {
		ctx.ReportNode(n, message)
		return
	}
	options, known := semanticquery.NativeArrayEntries(ctx, a)
	if !known {
		return
	}
	trace := options["s:trace"]
	if trace == nil {
		ctx.ReportNode(n, message)
		return
	}
	v, k := semanticquery.NativeTruth(ctx, trace)
	if k && !v {
		ctx.ReportNode(n, message)
	}
}
