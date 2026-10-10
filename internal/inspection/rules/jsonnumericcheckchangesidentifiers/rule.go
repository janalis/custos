// Package jsonnumericcheckchangesidentifiers implements the native JsonNumericCheckChangesIdentifiers inspection.
package jsonnumericcheckchangesidentifiers

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Preserve identifier strings during JSON encoding."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "JsonNumericCheckChangesIdentifiers" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "json_encode") {
		return
	}
	flags := semanticquery.CallArgument(call.Args, 1, "flags")
	yes, known := semanticquery.NativeFlag(ctx, flags, 32)
	if !known || !yes {
		return
	}
	if formatted(ctx, semanticquery.CallArgument(call.Args, 0, "value"), 0) {
		ctx.ReportNode(flags, message)
	}
}

func formatted(ctx *analysis.Context, e syntax.Expr, depth int) bool {
	if depth > 16 {
		return false
	}
	if s, ok := semanticquery.NativeString(ctx, e); ok {
		if len(s) < 2 || (s[0] != '+' && s[0] != '0') {
			return false
		}
		digits := strings.TrimPrefix(s, "+")
		for _, c := range digits {
			if c < '0' || c > '9' {
				return false
			}
		}
		return true
	}
	a, known := semanticquery.NativeArrayEntries(ctx, e)
	if !known {
		return false
	}
	for _, item := range a {
		if formatted(ctx, item, depth+1) {
			return true
		}
	}
	return false
}
