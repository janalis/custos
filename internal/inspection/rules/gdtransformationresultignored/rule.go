// Package gdtransformationresultignored implements the native GdTransformationResultIgnored inspection.
package gdtransformationresultignored

import (
	"custos/internal/inspection/analysis"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Use the image returned by the transformation."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "GdTransformationResultIgnored" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	if ctx.PHP < phpversion.PHP55 {
		return
	}
	c := n.(*syntax.FuncCall)
	switch semanticquery.NativeBuiltinName(ctx, c) {
	case "imagepng", "imagejpeg", "imagegif", "imagewebp", "imageavif":
	default:
		return
	}
	e := semanticquery.CallArgument(c.Args, 0, "image")
	if !semanticquery.ExpansionDUnaliased(ctx, e, n) {
		return
	}
	if discardedTransformation(ctx, c, e) {
		ctx.ReportNode(n, message)
	}
}

// discardedTransformation reports whether a discarded imagescale or imagecrop
// statement on the same image dominates the encoder. The source image is read
// positionally or through GD's `image` parameter name.
func discardedTransformation(ctx *analysis.Context, at *syntax.FuncCall, image syntax.Expr) bool {
	scope := syntax.EnclosingVariableScope(at)
	found := false
	visit := func(n syntax.Node) bool {
		if found || n.Span().Start >= at.Span().Start || (n != scope && syntax.IsVariableScope(n)) {
			return false
		}
		c, ok := n.(*syntax.FuncCall)
		if !ok {
			return true
		}
		if _, discarded := c.Parent().(*syntax.ExprStmt); !discarded || !semanticquery.NativeDominates(c, at) {
			return true
		}
		switch semanticquery.NativeBuiltinName(ctx, c) {
		case "imagescale", "imagecrop":
			found = semanticquery.NativeSameValue(ctx, semanticquery.CallArgument(c.Args, 0, "image"), image)
		}
		return !found
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return found
}
