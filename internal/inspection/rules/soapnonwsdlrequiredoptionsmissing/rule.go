// Package soapnonwsdlrequiredoptionsmissing implements the native SoapNonWsdlRequiredOptionsMissing inspection.
package soapnonwsdlrequiredoptionsmissing

import (
	"strings"

	"custos/internal/inspection/analysis"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply location and uri in non-WSDL mode."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SoapNonWsdlRequiredOptionsMissing" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KNew} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.New)
	cl := semanticquery.NativeNewClass(ctx, c)
	if cl == nil || !strings.HasPrefix(cl.File, "stubs/") || !strings.EqualFold(cl.FQN, "SoapClient") {
		return
	}
	if !syntax.IsNullConst(semanticquery.NativeLocalValue(ctx, semanticquery.CallArgument(c.Args, 0, "wsdl"))) {
		return
	}
	options, known := semanticquery.NativeArrayEntries(ctx, semanticquery.CallArgument(c.Args, 1, "options"))
	if known && (options["s:uri"] == nil || options["s:location"] == nil) {
		ctx.ReportNode(n, message)
	}
}
