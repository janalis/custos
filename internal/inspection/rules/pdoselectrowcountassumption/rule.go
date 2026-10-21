// Package pdoselectrowcountassumption implements the native PdoSelectRowCountAssumption inspection.
package pdoselectrowcountassumption

import (
	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use an explicit count query for portable row counts."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PdoSelectRowCountAssumption" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	check(ctx, n, message)
}

func check(ctx *analysis.Context, n syntax.Node, message string) {
	c := n.(*syntax.MethodCall)
	var fixes []diagnostic.Fix
	report := false

	if semanticquery.NativeMethod(ctx, c, "PDOStatement", "rowCount") {
		facts, known := semanticquery.PreparedSQL(ctx, c.Var)
		report = known && facts.SelectQuery
	}

	if report {
		ctx.ReportNode(c, message, fixes...)
	}
}
