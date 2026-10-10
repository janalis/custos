// Package pregsplitcaptureflagwithoutgroups implements the PregSplitCaptureFlagWithoutGroups inspection.
package pregsplitcaptureflagwithoutgroups

import (
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PregSplitCaptureFlagWithoutGroups" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

const message = "Remove the capture flag or add a capturing group."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call, _ := semanticquery.GlobalCall(ctx, n, "preg_split")
	if call == nil {
		return
	}
	pattern, ok := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 0, "pattern"))
	if !ok {
		return
	}
	count, known := semanticquery.NativeRegexCaptures(pattern)
	if !known || count != 0 {
		return
	}
	arg := semanticquery.CallArgument(call.Args, 3, "flags")
	has, known := semanticquery.NativeFlagContains(ctx, arg, "PREG_SPLIT_DELIM_CAPTURE")
	if !known || !has {
		return
	}
	if strings.Contains(ctx.Text(arg), "/*") || strings.Contains(ctx.Text(arg), "//") || strings.Contains(ctx.Text(arg), "#") {
		ctx.ReportNode(call, message)
		return
	}
	ctx.ReportNode(call, message, astquery.ReplaceFix(arg.Span(), withoutFlag(ctx, arg)))
}

func withoutFlag(ctx *analysis.Context, e syntax.Expr) string {
	e = syntax.UnwrapParens(e)
	if c, ok := e.(*syntax.ConstFetch); ok {
		if semanticquery.GlobalConstName(ctx, c) == "PREG_SPLIT_DELIM_CAPTURE" {
			return "0"
		}
		return ctx.Text(c)
	}
	if b, ok := e.(*syntax.Binary); ok && b.Op.Kind == syntax.TBar {
		left, right := withoutFlag(ctx, b.Left), withoutFlag(ctx, b.Right)
		if left == "0" {
			return right
		}
		if right == "0" {
			return left
		}
		return "(" + left + " | " + right + ")"
	}
	value, _ := semanticquery.NativeInt(ctx, e)
	return strconv.FormatInt(value&^2, 10)
}
func (rule) Semantic() {}
