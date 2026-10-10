// Package reflectionnamedargumentmismatch implements the native ReflectionNamedArgumentMismatch inspection.
package reflectionnamedargumentmismatch

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Use a declared reflection parameter name."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ReflectionNamedArgumentMismatch" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.MethodCall)
	if ctx.PHP < phpversion.PHP80 || (!semanticquery.NativeMethod(ctx, call, "ReflectionFunction", "invokeArgs") && !semanticquery.NativeMethod(ctx, call, "ReflectionMethod", "invokeArgs")) {
		return
	}
	params, _, known := semanticquery.NativeReflectionTarget(ctx, call.Var)
	if !known {
		return
	}
	for _, p := range params {
		if p.Variadic {
			return
		}
	}
	pos := 0
	if semanticquery.NativeConstruction(ctx, call.Var, "ReflectionMethod") != nil {
		pos = 1
	}
	entries, known := semanticquery.NativeArrayEntries(ctx, semanticquery.CallArgument(call.Args, pos, "args"))
	if !known {
		return
	}
	for key := range entries {
		if !strings.HasPrefix(key, "s:") {
			continue
		}
		found := false
		for _, p := range params {
			if strings.TrimPrefix(p.Name, "$") == key[2:] {
				found = true
			}
		}
		if !found {
			ctx.ReportNode(call, message)
			return
		}
	}
}
