// Package collatorsortdiscardsrequiredkeys implements the native CollatorSortDiscardsRequiredKeys inspection.
package collatorsortdiscardsrequiredkeys

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Preserve keys during collation sorting."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CollatorSortDiscardsRequiredKeys" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "Collator", "sort") {
		return
	}
	entries, ok := semanticquery.NativeArrayEntries(ctx, semanticquery.CallArgument(c.Args, 0, "array"))
	if !ok {
		return
	}
	for key := range entries {
		if strings.HasPrefix(key, "s:") {
			ctx.ReportNode(c, message, astquery.ReplaceFix(c.Name.Span(), "asort"))
			return
		}
	}
}
