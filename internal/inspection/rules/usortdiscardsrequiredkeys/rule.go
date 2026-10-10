// Package usortdiscardsrequiredkeys implements the native UsortDiscardsRequiredKeys inspection.
package usortdiscardsrequiredkeys

import (
	"strconv"
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Preserve record keys with uasort."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "UsortDiscardsRequiredKeys" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "usort") {
		return
	}
	input := semanticquery.CallArgument(call.Args, 0, "array")
	entries, known := semanticquery.NativeArrayEntries(ctx, input)
	if !known {
		return
	}
	st, ok := call.Parent().(*syntax.ExprStmt)
	if !ok {
		return
	}
	next, ok := astquery.NextStmt(ctx.File, st)
	if !ok {
		return
	}
	var lost bool
	syntax.Inspect(next, func(x syntax.Node) bool {
		if f, ok := x.(*syntax.ArrayDimFetch); ok && astquery.Equivalent(ctx.File, input, f.Var) {
			for parent := f.Parent(); parent != nil && parent != next.Parent(); parent = parent.Parent() {
				switch p := parent.(type) {
				case *syntax.Unset, *syntax.Isset, *syntax.Empty:
					return true
				case *syntax.Binary:
					if p.Op.Kind == syntax.TCoalesce && p.Left.Span().Contains(f.Span()) {
						return true
					}
				}

				if assignment, ok := parent.(*syntax.Assign); ok && assignment.Op.Kind == syntax.TEqual && assignment.Var.Span().Contains(f.Span()) {
					return true
				}
			}
			key, known := semanticquery.NativeArrayKey(ctx, f.Dim)
			if !known {
				return false
			}
			if _, exists := entries[key]; exists {
				index, err := strconv.ParseInt(strings.TrimPrefix(key, "i:"), 10, 64)
				if !strings.HasPrefix(key, "i:") || err != nil || index < 0 || index >= int64(len(entries)) {
					lost = true
				}
			}
		}
		switch x.(type) {
		case *syntax.Closure, *syntax.ArrowFunction, *syntax.Function, *syntax.ClassLike:
			return false
		}
		return !lost
	})
	if lost {
		name := call.Name.(*syntax.Name)
		ctx.ReportNode(call, message, diagnostic.Fix{Title: "Preserve array keys", Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: name.Span(), NewText: "\\uasort"}}
		}})
	}
}
