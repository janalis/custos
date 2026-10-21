// Package curlheaderoptionrequiresarray implements CurlHeaderOptionRequiresArray.
package curlheaderoptionrequiresarray

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply headers as an array."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CurlHeaderOptionRequiresArray" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	value := semanticquery.CurlSetting(ctx, c, "CURLOPT_HTTPHEADER")
	if value == nil {
		return
	}
	v := semanticquery.NativeLocalValue(ctx, value)
	switch v.(type) {
	case *syntax.Literal, *syntax.ConstFetch:
		if t := ctx.TypeOf(value); t.OnlyOf("string", "int", "float", "bool", "null") {
			ctx.ReportNode(c, message)
		}
	}
}
