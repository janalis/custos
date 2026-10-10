// Package httperrorbodyreadwithoutignoreerrors implements the native HttpErrorBodyReadWithoutIgnoreErrors inspection.
package httperrorbodyreadwithoutignoreerrors

import (
	"net/url"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Enable reading HTTP error response bodies."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "HttpErrorBodyReadWithoutIgnoreErrors" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)
	if name != "file_get_contents" && name != "fopen" {
		return
	}
	path, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "filename"))
	if !known {
		return
	}
	u, err := url.Parse(path)
	if err != nil || (strings.ToLower(u.Scheme) != "http" && strings.ToLower(u.Scheme) != "https") || u.Host == "" {
		return
	}
	slot := 2
	if name == "fopen" {
		slot = 3
	}
	context := semanticquery.CallArgument(c.Args, slot, "context")
	if context == nil {
		ctx.ReportNode(c, message)
		return
	}
	producer, ok := semanticquery.NativeValue(ctx, context).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, producer, "stream_context_create") {
		return
	}
	for _, call := range ctx.Flow().Calls(syntax.EnclosingVariableScope(c)) {
		if call.Node.Span().Start >= producer.Span().End && call.Node.Span().Start < c.Span().Start {
			return
		}
	}
	options := semanticquery.CallArgument(producer.Args, 0, "options")
	if options == nil {
		ctx.ReportNode(c, message)
		return
	}
	entries, known := semanticquery.NativeArrayEntries(ctx, options)
	if !known {
		return
	}
	http := entries["s:http"]
	if http == nil {
		ctx.ReportNode(c, message)
		return
	}
	fields, known := semanticquery.NativeArrayEntries(ctx, http)
	if !known {
		return
	}
	ignore := fields["s:ignore_errors"]
	if ignore == nil {
		ctx.ReportNode(c, message)
		return
	}
	enabled, known := semanticquery.NativeTruth(ctx, ignore)
	if known && !enabled {
		ctx.ReportNode(c, message)
	}
}
