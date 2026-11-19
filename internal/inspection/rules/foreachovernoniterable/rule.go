// Package foreachovernoniterable implements the native
// ForeachOverNonIterable inspection.
package foreachovernoniterable

import (
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// New constructs the stateless inspection.
func New() analysis.Rule { return rule{} }

type rule struct{}

func (rule) ID() string { return "ForeachOverNonIterable" }

func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KForeach} }

func (rule) Semantic() {}

// Check reports the subject of a loop when native declarations prove it
// holds only scalars or null.
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	subject := n.(*syntax.Foreach).Expr
	typ := ctx.Types().ProvenTypeOf(subject)
	if typ.IsUnknown() {
		return
	}
	for _, a := range typ.Atoms() {
		switch a {
		case "int", "float", "string", "bool", "true", "false", "null":
		default:
			return
		}
	}
	ctx.ReportNode(subject, "Iterate over an array or an object, not "+typ.String()+".")
}
