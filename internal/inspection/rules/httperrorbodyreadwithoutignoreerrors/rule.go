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
		if !defaultContextTouched(ctx) {
			ctx.ReportNode(c, message)
		}
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
	// The params array may carry its own "options" entry, merged into the context.
	if params, state := semanticquery.CallArgumentBinding(producer.Args, 1, "params"); state != semanticquery.ArgumentMissing {
		entries, known := semanticquery.NativeArrayEntries(ctx, params)
		if !known || entries["s:options"] != nil {
			return
		}
	}
	options, state := semanticquery.CallArgumentBinding(producer.Args, 0, "options")
	if state == semanticquery.ArgumentMissing {
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

// defaultContextTouched reports whether the file configures or exposes the
// process-wide default stream context, which a context-less read then uses.
func defaultContextTouched(ctx *analysis.Context) bool {
	touched := false
	syntax.InspectFile(ctx.File, func(n syntax.Node) bool {
		if call, ok := n.(*syntax.FuncCall); ok {
			name := semanticquery.NativeBuiltinName(ctx, call)
			touched = touched || name == "stream_context_set_default" || name == "stream_context_get_default"
		}
		return !touched
	})
	return touched
}
