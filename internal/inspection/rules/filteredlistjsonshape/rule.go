// Package filteredlistjsonshape implements the native FilteredListJsonShape inspection.
package filteredlistjsonshape

import (
	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Reindex the filtered array before encoding a JSON list."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FilteredListJsonShape" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "json_encode") {
		return
	}
	flags := semanticquery.CallArgument(call.Args, 1, "flags")
	forced, known := semanticquery.NativeFlag(ctx, flags, 16) // JSON_FORCE_OBJECT
	if !known || forced {
		return
	}
	e := semanticquery.CallArgument(call.Args, 0, "value")
	inner, ok := syntax.UnwrapParens(e).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, inner, "array_filter") {
		return
	}
	callback := semanticquery.CallArgument(inner.Args, 1, "callback")
	if callback != nil && !ctx.TypeOf(callback).OnlyOf("null") {
		return
	}
	source := semanticquery.CallArgument(inner.Args, 0, "array")
	a := semanticquery.NativeArray(ctx, source)
	if a == nil {
		return
	}
	gap, broken := false, false
	for i, item := range a.Items {
		if item.Key != nil {
			key, known := semanticquery.NativeInt(ctx, item.Key)
			if !known || key != int64(i) {
				return
			}
		}
		truth, known := semanticquery.NativeTruth(ctx, item.Value)
		if !known {
			return
		}
		if !truth {
			gap = true
		} else if gap {
			broken = true
		}
	}
	scope := syntax.EnclosingFuncLike(call)
	if broken && !reboundOutOfBand(ctx.File, scope, source, map[string]bool{}) && !astquery.CallsExtract(ctx.File, scope) {
		ctx.ReportNode(call, message, diagnostic.Fix{Title: "Reindex the filtered list", Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: inner.Span(), NewText: "\\array_values(" + ctx.Text(inner) + ")"}}
		}})
	}
}

// reboundOutOfBand reports whether a local the filtered list is resolved
// through (the array variable, a variable item or key, or a variable assigned
// to one of those) can receive another value without a plain assignment in
// the scope: through a reference to it (a by-reference closure capture, array
// item, assignment source or foreach subject) or a binding that is not an
// assignment (`++`, `--`, catch, foreach target, global, static, unset).
// The resolved literal is then not proven to be what is filtered.
func reboundOutOfBand(f *syntax.File, scope syntax.Node, e syntax.Expr, seen map[string]bool) bool {
	rebound := false
	syntax.Inspect(e, func(x syntax.Node) bool {
		v, ok := x.(*syntax.Variable)
		if !ok || rebound || seen[v.Name] {
			return !rebound
		}
		seen[v.Name] = true
		for _, access := range flowquery.VarAccesses(f, scope, v.Name) {
			switch by := access.By.(type) {
			case nil:
				switch parent := access.Var.Parent().(type) {
				case *syntax.ArrayItem:
					rebound = rebound || parent.ByRef
				case *syntax.Assign:
					rebound = rebound || parent.ByRef
				case *syntax.Foreach:
					rebound = rebound || parent.ByRef
				}
			case *syntax.Assign:
				rebound = rebound || by.ByRef || reboundOutOfBand(f, scope, by.Value, seen)
			case *syntax.ClosureUse:
				rebound = rebound || by.ByRef
			default:
				rebound = true
			}
		}
		return false
	})
	return rebound
}
