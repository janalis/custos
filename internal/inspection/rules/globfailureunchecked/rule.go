// Package globfailureunchecked implements the native GlobFailureUnchecked inspection.
package globfailureunchecked

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

const (
	message      = "Handle glob failure before iterating its result."
	arrayMessage = "Handle glob failure before using its result as an array."
)

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule { return rule{} }
func (rule) ID() string  { return "GlobFailureUnchecked" }
func (rule) Semantic()   {}
func (rule) Flow()       {}
func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KForeach, syntax.KFuncCall}
}

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	text := message
	if _, ok := n.(*syntax.FuncCall); ok {
		text = arrayMessage
	}
	check(ctx, n, text)
}

func check(ctx *analysis.Context, n syntax.Node, message string) {
	switch x := n.(type) {
	case *syntax.Foreach:
		globFailureUse(ctx, x.Expr, message)
	case *syntax.FuncCall:
		// Cheap lexical filter first: almost no call is an array-only consumer.
		name, ok := x.Name.(*syntax.Name)
		if !ok {
			return
		}
		consumer, ok := globArrayConsumers[strings.ToLower(name.Value[strings.LastIndexByte(name.Value, '\\')+1:])]
		if ok && semanticquery.NativeBuiltinName(ctx, x) != "" {
			globFailureUse(ctx, semanticquery.CallArgument(x.Args, consumer.position, consumer.name), message)
		}
	}
}

// globArrayConsumers lists builtins whose array parameter rejects false: a
// TypeError since PHP 8.0, a warning and a meaningless result before.
var globArrayConsumers = map[string]struct {
	position int
	name     string
}{
	"count": {0, "value"}, "sizeof": {0, "value"},
	"array_map": {1, "array"}, "array_filter": {0, "array"}, "array_values": {0, "array"},
	"array_keys": {0, "array"}, "array_reverse": {0, "array"}, "array_slice": {0, "array"},
	"array_unique": {0, "array"}, "array_flip": {0, "array"}, "array_sum": {0, "array"},
	"array_product": {0, "array"}, "array_column": {0, "array"}, "array_chunk": {0, "array"},
	"array_key_first": {0, "array"}, "array_key_last": {0, "array"}, "array_is_list": {0, "array"},
	"array_rand": {0, "array"}, "array_key_exists": {1, "array"}, "in_array": {1, "haystack"},
	"array_search": {1, "haystack"}, "implode": {1, "array"}, "join": {1, "array"},
	"array_combine": {0, "keys"}, "array_fill_keys": {0, "keys"},
	// Variadic array lists cannot be named; only the first array is inspected.
	"array_merge": {0, ""}, "array_diff": {0, ""}, "array_intersect": {0, ""},
}

// globFailureUse reports an array-only use whose single reaching value is a
// builtin glob() call that no flow fact or recognised guard separates from false.
func globFailureUse(ctx *analysis.Context, use syntax.Expr, message string) {
	call, ok := semanticquery.NativeLocalBuiltinResult(ctx, use, "false", "glob").(*syntax.FuncCall)
	if !ok || semanticquery.NativeBuiltinName(ctx, call) != "glob" || ctx.Flow().Excludes(use, "false") || !semanticquery.NativeFailureFlowOnly(ctx, use) && semanticquery.NativeNarrowedExcludes(ctx, use, types.Of("array", "false"), "false") {
		return
	}
	ctx.ReportNode(use, message)
}
