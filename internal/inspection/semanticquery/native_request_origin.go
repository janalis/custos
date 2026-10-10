package semanticquery

import (
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

type (
	nativeRequestWritesKey struct{ scope syntax.Node }
	nativeRequestWrites    struct {
		first    map[string]uint32
		complete bool
	}
)

// NativeRequestUnwritten excludes explicit mutations and reference escapes of
// a request superglobal before the source read. Only syntax facts are cached.
func NativeRequestUnwritten(ctx *analysis.Context, source *syntax.ArrayDimFetch) bool {
	root, ok := source.Var.(*syntax.Variable)
	if !ok {
		return false
	}
	scope := syntax.EnclosingVariableScope(source)
	facts := ctx.File.Memo(nativeRequestWritesKey{scope}, func() any {
		facts := nativeRequestWrites{first: map[string]uint32{}, complete: true}
		count := 0
		visit := func(n syntax.Node) bool {
			count++
			if count > 4096 {
				facts.complete = false
				return false
			}
			if n != scope && syntax.IsVariableScope(n) {
				return false
			}
			var targets []syntax.Expr
			switch x := n.(type) {
			case *syntax.New:
				if first, known := facts.first["@constructor"]; !known || x.Span().Start < first {
					facts.first["@constructor"] = x.Span().Start
				}
			case *syntax.Assign:
				targets = append(targets, x.Var)
				if x.ByRef {
					targets = append(targets, x.Value)
				}
			case *syntax.IncDec:
				targets = append(targets, x.Var)
			case *syntax.Unset:
				targets = x.Vars
			}
			for _, target := range targets {
				syntax.Inspect(target, func(n syntax.Node) bool {
					if v, ok := n.(*syntax.Variable); ok {
						if first, known := facts.first[v.Name]; !known || v.Span().Start < first {
							facts.first[v.Name] = v.Span().Start
						}
					}
					return true
				})
			}
			return true
		}
		if scope == nil {
			syntax.InspectFile(ctx.File, visit)
		} else {
			syntax.Inspect(scope, visit)
		}
		return facts
	}).(nativeRequestWrites)
	if !facts.complete {
		return false
	}
	if first, known := facts.first["@constructor"]; known && first < source.Span().Start {
		return false
	}
	if first, known := facts.first["GLOBALS"]; known && first < source.Span().Start {
		return false
	}
	if first, known := facts.first[root.Name]; known && first < source.Span().Start {
		return false
	}
	for _, call := range ctx.Flow().Calls(scope) {
		if call.Node.Span().End > source.Span().Start {
			continue
		}
		c, ok := call.Node.(*syntax.FuncCall)
		if !ok {
			return false
		}
		switch NativeBuiltinName(ctx, c) {
		case "strlen", "is_string", "ctype_alnum", "ctype_digit", "in_array":
		default:
			return false
		}
	}
	return true
}
