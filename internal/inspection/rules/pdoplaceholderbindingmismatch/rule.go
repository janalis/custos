// Package pdoplaceholderbindingmismatch implements the native PdoPlaceholderBindingMismatch inspection.
package pdoplaceholderbindingmismatch

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Match SQL placeholders to the effective parameter bindings."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PdoPlaceholderBindingMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "PDOStatement", "execute") {
		return
	}
	facts, known := semanticquery.PreparedSQL(ctx, c.Var)
	if !known {
		return
	}

	bindings, known := semanticquery.ExecuteBindings(ctx, c)
	if !known {
		return
	}
	report := len(bindings) != len(facts.Markers)
	for key := range facts.Markers {
		if !bindings[key] {
			report = true
		}
	}

	if report {
		ctx.ReportNode(c, message)
	}
}
