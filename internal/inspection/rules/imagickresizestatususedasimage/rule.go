// Package imagickresizestatususedasimage implements the native ImagickResizeStatusUsedAsImage inspection.
package imagickresizestatususedasimage

import (
	"custos/internal/inspection/analysis"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use the mutated Imagick object after resizing."

type rule struct{}

func New() analysis.Rule { return rule{} }
func (rule) ID() string  { return "ImagickResizeStatusUsedAsImage" }
func (rule) Semantic()   {}
func (rule) Flow()       {}
func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KMethodCall, syntax.KPropertyFetch}
}

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	var e syntax.Expr
	switch c := n.(type) {
	case *syntax.MethodCall:
		e = c.Var
	case *syntax.PropertyFetch:
		e = c.Var
	}
	c, ok := semanticquery.NativeLocalValue(ctx, e).(*syntax.MethodCall)
	if ok && semanticquery.NativeMethod(ctx, c, "Imagick", "resizeImage") {
		ctx.ReportNode(n, message)
	}
}
