// Package compressionleveloutsidesupportedrange implements the native CompressionLevelOutsideSupportedRange inspection.
package compressionleveloutsidesupportedrange

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/flowquery"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
)

const message = "Use a supported compression level."

type rule struct{}

// New returns the CompressionLevelOutsideSupportedRange inspection.
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
	arg := semanticquery.CallArgument(c.Args, 1, "level")
	level, ok := semanticquery.NativeInt(ctx, arg)
	if !ok || level >= -1 && level <= 9 {
		return
	}
	if !semanticquery.StableLocalValue(ctx, arg, semanticquery.LocalStability{SharedOpaque: true}) {
		return
	}
	ctx.ReportNode(c, message)
}
