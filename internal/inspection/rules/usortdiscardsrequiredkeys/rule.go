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
	// A read proves the key is required only when nothing in the same statement
	// tests the key for presence, and nothing earlier in it writes the key or
	// replaces the whole array.
	type read struct {
		key string
		at  uint32
	}
	var reads []read
	tested := map[string]bool{}
	written := map[string]uint32{}
	write := func(key string, end uint32) {
		if old, seen := written[key]; !seen || end < old {
			written[key] = end
		}
	}
	syntax.Inspect(next, func(x syntax.Node) bool {
		switch x := x.(type) {
		case *syntax.Closure, *syntax.ArrowFunction, *syntax.Function, *syntax.ClassLike:
			return false
		case *syntax.Assign:
			if astquery.Equivalent(ctx.File, input, x.Var) {
				write("", x.Span().End)
			}
		case *syntax.FuncCall:
			if !semanticquery.NativeBuiltin(ctx, x, "array_key_exists") && !semanticquery.NativeBuiltin(ctx, x, "key_exists") {
				return true
			}
			if astquery.Equivalent(ctx.File, input, semanticquery.CallArgument(x.Args, 1, "array")) {
				if key, known := semanticquery.NativeArrayKey(ctx, semanticquery.CallArgument(x.Args, 0, "key")); known {
					tested[key] = true
				}
			}
		case *syntax.ArrayDimFetch:
			if x.Dim == nil || !astquery.Equivalent(ctx.File, input, x.Var) {
				return true
			}
			key, known := semanticquery.NativeArrayKey(ctx, x.Dim)
			if !known {
				return true
			}
			for parent := x.Parent(); parent != nil && parent != next.Parent(); parent = parent.Parent() {
				switch p := parent.(type) {
				case *syntax.Isset, *syntax.Empty:
					tested[key] = true
					return true
				case *syntax.Unset:
					write(key, p.Span().End)
					return true
				case *syntax.Binary:
					if p.Op.Kind == syntax.TCoalesce && p.Left.Span().Contains(x.Span()) {
						tested[key] = true
						return true
					}
				case *syntax.Assign:
					if !p.Var.Span().Contains(x.Span()) {
						continue
					}
					switch p.Op.Kind {
					case syntax.TCoalesceEqual:
						tested[key] = true
						return true
					case syntax.TEqual:
						write(key, p.Span().End)
						return true
					}
				}
			}
			if _, exists := entries[key]; exists {
				index, err := strconv.ParseInt(strings.TrimPrefix(key, "i:"), 10, 64)
				if !strings.HasPrefix(key, "i:") || err != nil || index < 0 || index >= int64(len(entries)) {
					reads = append(reads, read{key, x.Span().Start})
				}
			}
		}
		return true
	})
	var lost bool
	for _, r := range reads {
		replaced, whole := written[""]
		end, partial := written[r.key]
		if !tested[r.key] && (!whole || r.at < replaced) && (!partial || r.at < end) {
			lost = true
		}
	}
	if lost {
		name := call.Name.(*syntax.Name)
		ctx.ReportNode(call, message, diagnostic.Fix{Title: "Preserve array keys", Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: name.Span(), NewText: "\\uasort"}}
		}})
	}
}
