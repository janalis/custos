// Package numberparseusescurrencytype implements the native NumberParseUsesCurrencyType inspection.
package numberparseusescurrencytype

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/flowquery"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
)

const message = "Parse currency with parseCurrency."

type rule struct{}

// New returns the NumberParseUsesCurrencyType inspection.
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
	if semanticquery.NativeMethod(ctx, c, "NumberFormatter", "parse") && semanticquery.BuiltinClassConstant(ctx, semanticquery.CallArgument(c.Args, 1, "type"), "NumberFormatter", "TYPE_CURRENCY") {
		ctx.ReportNode(c, message)
	}
}
