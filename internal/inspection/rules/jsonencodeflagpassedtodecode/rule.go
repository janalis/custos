// Package jsonencodeflagpassedtodecode implements the native JsonEncodeFlagPassedToDecode inspection.
package jsonencodeflagpassedtodecode

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Use decoding flags in json_decode."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "JsonEncodeFlagPassedToDecode" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP54 {
		return
	}
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "json_decode") {
		return
	}
	flags := semanticquery.CallArgument(call.Args, 3, "flags")
	if encodingFlag(ctx, flags, 0) {
		ctx.ReportNode(call, message)
	}
}

func encodingFlag(ctx *analysis.Context, e syntax.Expr, depth int) bool {
	if depth > 16 {
		return false
	}
	e = syntax.UnwrapParens(e)
	if b, ok := e.(*syntax.Binary); ok && b.Op.Kind == syntax.TBar {
		return encodingFlag(ctx, b.Left, depth+1) || encodingFlag(ctx, b.Right, depth+1)
	}
	c, ok := e.(*syntax.ConstFetch)
	if !ok {
		return false
	}
	switch semanticquery.GlobalConstName(ctx, c) {
	case "JSON_HEX_TAG", "JSON_HEX_AMP", "JSON_HEX_APOS", "JSON_HEX_QUOT", "JSON_FORCE_OBJECT", "JSON_NUMERIC_CHECK", "JSON_UNESCAPED_SLASHES", "JSON_PRETTY_PRINT", "JSON_UNESCAPED_UNICODE", "JSON_PARTIAL_OUTPUT_ON_ERROR", "JSON_PRESERVE_ZERO_FRACTION":
		return true
	}
	return false
}
