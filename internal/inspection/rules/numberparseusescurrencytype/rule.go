// Package numberparseusescurrencytype implements the native NumberParseUsesCurrencyType inspection.
package numberparseusescurrencytype

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Parse currency with parseCurrency."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "NumberParseUsesCurrencyType" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.MethodCall)
	if semanticquery.NativeMethod(ctx, c, "NumberFormatter", "parse") && semanticquery.ExpansionCConstant(ctx, semanticquery.CallArgument(c.Args, 1, "type"), "NumberFormatter", "TYPE_CURRENCY") {
		ctx.ReportNode(c, message)
	}
}
