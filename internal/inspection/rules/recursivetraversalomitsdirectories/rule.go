// Package recursivetraversalomitsdirectories implements the native RecursiveTraversalOmitsDirectories inspection.
package recursivetraversalomitsdirectories

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Include parent nodes when traversing directories."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "RecursiveTraversalOmitsDirectories" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KNew} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.New)
	if semanticquery.NativeConstruction(ctx, c, "RecursiveIteratorIterator") == nil {
		return
	}
	source := semanticquery.CallArgument(c.Args, 0, "iterator")
	if wrapper := semanticquery.NativeConstruction(ctx, source, "RecursiveCallbackFilterIterator"); wrapper != nil {
		source = semanticquery.CallArgument(wrapper.Args, 0, "iterator")
	}
	if semanticquery.NativeConstruction(ctx, source, "RecursiveDirectoryIterator") == nil {
		return
	}
	mode := semanticquery.CallArgument(c.Args, 1, "mode")
	if mode != nil {
		v, known := semanticquery.NativeContractInt(ctx, mode)
		if !known || v != 0 {
			return
		}
	}
	ctx.ReportNode(c, message)
}
