// Package pdoquotedplaceholder implements the native PdoQuotedPlaceholder inspection.
package pdoquotedplaceholder

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
)

const message = "Remove SQL quotes around this parameter marker."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PdoQuotedPlaceholder" }
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
	for key := range bindings {
		if facts.Quoted[key] && !facts.Markers[key] {
			ctx.ReportNode(c, message)
			return
		}
	}
}
