// Package trimmaskusedassuffix implements the TrimMaskUsedAsSuffix inspection.
package trimmaskusedassuffix

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "TrimMaskUsedAsSuffix" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

const message = "Remove the suffix explicitly instead of using a trim mask."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call, _ := semanticquery.GlobalCall(ctx, n, "trim", "ltrim", "rtrim")
	if call == nil {
		return
	}
	mask, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 1, "characters"))
	if !known || len(mask) < 3 || mask[0] != '.' {
		return
	}
	for _, c := range mask[1:] {
		if c < 'a' || c > 'z' {
			if c < 'A' || c > 'Z' {
				return
			}
		}
	}
	ctx.ReportNode(call, message)
}

func (rule) Semantic() {}
