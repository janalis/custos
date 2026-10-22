// Package curluploaddeclaredsizemismatch implements CurlUploadDeclaredSizeMismatch.
package curluploaddeclaredsizemismatch

import (
	"net/url"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Match the declared upload size to remaining bytes."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CurlUploadDeclaredSizeMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	size := semanticquery.CurlSetting(ctx, c, "CURLOPT_INFILESIZE")
	if size == nil {
		return
	}
	wanted, known := semanticquery.NativeContractInt(ctx, size)
	if !known || wanted < 0 {
		return
	}
	handle := semanticquery.CallArgument(c.Args, 0, "handle")
	uploading, known := semanticquery.NativeTruth(ctx, curlOption(ctx, c, handle, "CURLOPT_UPLOAD"))
	if !known || !uploading {
		return
	}
	source := curlOption(ctx, c, handle, "CURLOPT_INFILE")
	open := semanticquery.OriginCall(ctx, source, "fopen")
	if open == nil {
		return
	}
	path, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(open.Args, 0, "filename"))
	if !known || !strings.HasPrefix(path, "data://text/plain,") {
		return
	}
	mode, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(open.Args, 1, "mode"))
	if !known || (mode != "r" && mode != "rb") {
		return
	}
	for _, event := range semanticquery.ScopeCallEvents(ctx, c) {
		p, ok := event.(*syntax.FuncCall)
		if !ok || p == open || p.Span().Start <= open.Span().End || p.Span().End >= c.Span().Start {
			continue
		}
		if semanticquery.SameStableLocal(ctx, semanticquery.CallArgument(p.Args, 0, "stream"), source) && !semanticquery.NativeBuiltin(ctx, p, "curl_setopt") {
			return
		}
	}
	data, err := url.PathUnescape(strings.TrimPrefix(path, "data://text/plain,"))
	if err == nil && int64(len(data)) != wanted {
		ctx.ReportNode(c, message)
	}
}

func curlOption(ctx *analysis.Context, at syntax.Node, handle syntax.Expr, option string) syntax.Expr {
	var result syntax.Expr
	for _, event := range semanticquery.ScopeCallEvents(ctx, at) {
		c, ok := event.(*syntax.FuncCall)
		if !ok || c.Span().End >= at.Span().Start || !semanticquery.SameStableLocal(ctx, semanticquery.CallArgument(c.Args, 0, "handle"), handle) {
			continue
		}
		if !semanticquery.NativeDominates(c, at) {
			return nil
		}
		if !semanticquery.NativeBuiltin(ctx, c, "curl_setopt") && !semanticquery.NativeBuiltin(ctx, c, "curl_setopt_array") {
			return nil
		}
		if value := semanticquery.CurlSetting(ctx, c, option); value != nil {
			result = value
		}
	}
	return result
}
