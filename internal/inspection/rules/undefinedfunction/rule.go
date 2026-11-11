// Package undefinedfunction implements the UndefinedFunction inspection.
package undefinedfunction

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// New constructs the stateless inspection.
func New() analysis.Rule { return rule{} }

type rule struct{}

func (rule) ID() string { return "UndefinedFunction" }

func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func (rule) Semantic() {}

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	name, ok := call.Name.(*syntax.Name)
	if !ok {
		return
	}
	fqn, fallback := ctx.Names().Function(name.Value, name.Span().Start)
	candidates := []string{fqn}
	if fallback != "" {
		candidates = append(candidates, fallback)
	}
	if !semanticquery.UndeclaredEverywhere(ctx, true, candidates...) || semanticquery.ExistenceChecked(ctx, call, candidates...) {
		return
	}
	// An unqualified name names the global function the call falls back to.
	shown := fqn
	if fallback != "" {
		shown = fallback
	}
	ctx.ReportNode(name, "Fix the name or install the package declaring "+shown+"().")
}
