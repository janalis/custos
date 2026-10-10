// Package jsonforceobjectchangesnestedlists implements the native JsonForceObjectChangesNestedLists inspection.
package jsonforceobjectchangesnestedlists

import (
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Preserve nested list representation during JSON encoding."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "JsonForceObjectChangesNestedLists" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "json_encode") {
		return
	}
	flags := semanticquery.CallArgument(call.Args, 1, "flags")
	yes, known := semanticquery.NativeFlag(ctx, flags, 16)
	if !known || !yes {
		return
	}
	if nested(ctx, semanticquery.CallArgument(call.Args, 0, "value"), false, 0) {
		ctx.ReportNode(flags, message)
	}
}

func nested(ctx *analysis.Context, e syntax.Expr, child bool, depth int) bool {
	if depth > 16 {
		return false
	}
	a, known := semanticquery.NativeArrayEntries(ctx, e)
	if !known {
		return false
	}
	list := len(a) > 0
	for key := range a {
		if !strings.HasPrefix(key, "i:") {
			list = false
			continue
		}
		n, _ := strconv.ParseInt(key[2:], 10, 64)
		if n < 0 || n >= int64(len(a)) {
			list = false
		}
	}
	if child && list {
		return true
	}
	for _, item := range a {
		if nested(ctx, item, true, depth+1) {
			return true
		}
	}
	return false
}
