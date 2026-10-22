// Package pdomixedplaceholderstyles implements the native PdoMixedPlaceholderStyles inspection.
package pdomixedplaceholderstyles

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use one SQL placeholder style per statement."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PdoMixedPlaceholderStyles" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "PDO", "prepare") {
		return
	}
	sql, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "query"))
	if !known {
		return
	}
	facts, known := semanticquery.ScanSQL(sql)
	if !known {
		return
	}
	if facts.Positional > 0 && len(facts.Markers) > facts.Positional {
		ctx.ReportNode(c, message)
	}
}
