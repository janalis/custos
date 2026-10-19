// Package datelasterrorsfalseunchecked implements the DateLastErrorsFalseUnchecked inspection.
package datelasterrorsfalseunchecked

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/types"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DateLastErrorsFalseUnchecked" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KArrayDimFetch} }

const message = "Handle the false result before reading parsing diagnostics."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP82 {
		return
	}
	fetch := n.(*syntax.ArrayDimFetch)
	origin := semanticquery.NativeValue(ctx, fetch.Var)
	if ctx.Flow().Excludes(fetch.Var, "false") {
		return
	}
	if semanticquery.NativeDateStatic(ctx, origin, "getLastErrors") == nil {
		return
	}
	if silentProbe(fetch) || !semanticquery.NativeLabelled(ctx, fetch.Var) && semanticquery.NativeNarrowedExcludes(ctx, fetch.Var, types.Of("array", "false"), "false") {
		return
	}
	ctx.ReportNode(fetch, message)
}
func (rule) Semantic() {}

// silentProbe recognises the contexts where PHP evaluates an offset of false
// without a diagnostic: isset, empty and the tested side of `??`.
func silentProbe(fetch *syntax.ArrayDimFetch) bool {
	for child, parent := syntax.Node(fetch), fetch.Parent(); ; child, parent = parent, parent.Parent() {
		switch p := parent.(type) {
		case *syntax.ArrayDimFetch:
			if p.Var != child {
				return false
			}
		case *syntax.Paren:
		case *syntax.Isset, *syntax.Empty:
			return true
		case *syntax.Binary:
			return p.Op.Kind == syntax.TCoalesce && p.Left == child
		default:
			return false
		}
	}
}
