// Package compressionleveloutsidesupportedrange implements the native CompressionLevelOutsideSupportedRange inspection.
package compressionleveloutsidesupportedrange

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use a supported compression level."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CompressionLevelOutsideSupportedRange" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	switch semanticquery.NativeBuiltinName(ctx, c) {
	case "gzencode", "gzcompress", "gzdeflate":
	default:
		return
	}
	level, ok := semanticquery.NativeInt(ctx, semanticquery.CallArgument(c.Args, 1, "level"))
	if ok && (level < -1 || level > 9) {
		ctx.ReportNode(c, message)
	}
}
