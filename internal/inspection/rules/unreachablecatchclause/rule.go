// Package unreachablecatchclause implements the native UnreachableCatchClause inspection.
package unreachablecatchclause

import (
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

const message = "Move the specific catch before the broader catch."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "UnreachableCatchClause" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KTry} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	tr := n.(*syntax.Try)
	for i, c := range tr.Catches {
		if i == 0 || len(c.Types) == 0 {
			continue
		}
		covered := true
		for _, typ := range c.Types {
			later := ctx.Names().Class(typ.Value, typ.Span().Start)
			found := false
			for _, earlier := range tr.Catches[:i] {
				for _, t := range earlier.Types {
					parent := ctx.Names().Class(t.Value, t.Span().Start)
					if ctx.Index().IsSubtype(later, parent, ctx.PHP) {
						found = true
					}
				}
			}
			if !found {
				covered = false
				break
			}
		}
		if covered {
			ctx.Report(syntax.Span{Start: c.Types[0].Span().Start, End: c.Types[len(c.Types)-1].Span().End}, message)
		}
	}
}
