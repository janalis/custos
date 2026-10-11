// Package splfixedarrayindexoutofbounds implements the native SplFixedArrayIndexOutOfBounds inspection.
package splfixedarrayindexoutofbounds

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use an index within the fixed array size."

// readers are builtins that inspect an object argument without keeping it,
// resizing it or handing it to user code.
var readers = map[string]bool{
	"count": true, "sizeof": true, "iterator_count": true, "iterator_to_array": true,
	"var_dump": true, "print_r": true, "var_export": true, "json_encode": true,
	"get_class": true, "is_object": true, "is_null": true, "spl_object_id": true, "spl_object_hash": true,
}

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SplFixedArrayIndexOutOfBounds" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KArrayDimFetch} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	f := n.(*syntax.ArrayDimFetch)
	i, ok := semanticquery.NativeInt(ctx, f.Dim)
	if !ok || probed(f) || !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	size, known := semanticquery.NativeFixedArraySize(ctx, f.Var, f)
	frozen := false
	if !known {
		// An element unset hides a factory allocation from the value graph;
		// recover it from the direct assignment and accept no later resize.
		origin, ok := semanticquery.NativeLocalValue(ctx, f.Var).(*syntax.StaticCall)
		if !ok {
			return
		}
		size, known = semanticquery.NativeFixedArraySize(ctx, origin, f)
		frozen = true
	}
	if known && (i < 0 || i >= size) && stable(ctx, f, frozen) {
		ctx.ReportNode(f, message)
	}
}

// probed reports whether the fetch is only tested for existence: isset(),
// empty() and the left operand of ?? read in a mode that never throws, also
// through a longer offset or property chain based on the fetch.
func probed(f *syntax.ArrayDimFetch) bool {
	var n syntax.Node = f
	for {
		switch p := n.Parent().(type) {
		case *syntax.Paren:
			n = p
		case *syntax.ArrayDimFetch:
			if p.Var != n {
				return false
			}
			n = p
		case *syntax.PropertyFetch:
			if p.Var != n {
				return false
			}
			n = p
		case *syntax.Isset, *syntax.Empty:
			return true
		case *syntax.Binary:
			return p.Op.Kind == syntax.TCoalesce && p.Left == n
		default:
			return false
		}
	}
}

// stable proves that nothing but the tracked assignments and dominating
// setSize calls can have changed the receiver before the access: every other
// mention must be inert. Inside a loop the whole loop counts as "before".
func stable(ctx *analysis.Context, f *syntax.ArrayDimFetch, frozen bool) bool {
	recv := syntax.UnwrapParens(f.Var)
	_, local := recv.(*syntax.Variable)
	scope := syntax.EnclosingVariableScope(f)
	limit := f.Span().Start
	for p := f.Parent(); p != nil && p != scope; p = p.Parent() {
		switch p.(type) {
		case *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach:
			limit = p.Span().End
		}
	}
	ok := true
	visit := func(n syntax.Node) bool {
		if !ok || n.Span().Start >= limit {
			return false
		}
		if n != scope && syntax.IsVariableScope(n) {
			ok = !captures(ctx, n, recv, local)
			return false
		}
		if astquery.Equivalent(ctx.File, n, recv) {
			ok = inert(ctx, n, f, frozen)
			return false
		}
		// A property outlives the scope: any unknown code may resize it.
		if !local {
			ok = !unknownCall(ctx, n, recv)
		}
		return ok
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return ok
}

// captures reports whether a nested closure or arrow function can reach the
// receiver; other nested declarations have their own unrelated locals.
func captures(ctx *analysis.Context, scope syntax.Node, recv syntax.Expr, local bool) bool {
	switch c := scope.(type) {
	case *syntax.Closure:
		if local {
			for _, use := range c.Uses {
				if astquery.Equivalent(ctx.File, use.Var, recv) {
					return true
				}
			}
			return false
		}
	case *syntax.ArrowFunction:
	default:
		return false
	}
	found := false
	syntax.Inspect(scope, func(n syntax.Node) bool {
		found = found || astquery.Equivalent(ctx.File, n, recv)
		return !found
	})
	return found
}

// inert accepts the mentions of the receiver that neither resize it nor let
// it escape: offset access, builtin method calls, its own assignment,
// by-value iteration, existence tests, cloning and read-only builtin
// arguments.
func inert(ctx *analysis.Context, n syntax.Node, f *syntax.ArrayDimFetch, frozen bool) bool {
	switch p := n.Parent().(type) {
	case *syntax.ArrayDimFetch:
		return p.Var == n
	case *syntax.MethodCall:
		id, named := p.Name.(*syntax.Identifier)
		if p.Var != n || !named {
			return false
		}
		return !strings.EqualFold(id.Value, "setSize") || (!frozen && semanticquery.NativeDominates(p, f))
	case *syntax.Assign:
		return p.Var == n && !p.ByRef && p.Span().Start < f.Span().Start
	case *syntax.Foreach:
		return p.Expr == n && !p.ByRef
	case *syntax.Isset, *syntax.Empty, *syntax.Clone:
		return true
	case *syntax.Arg:
		call, ok := p.Parent().Parent().(*syntax.FuncCall)
		return ok && !p.Unpack && !p.ByRef && readers[strings.ToLower(semanticquery.NativeBuiltinName(ctx, call))]
	}
	return false
}

// unknownCall reports code that may run user logic able to reach a property.
func unknownCall(ctx *analysis.Context, n syntax.Node, recv syntax.Expr) bool {
	switch c := n.(type) {
	case *syntax.MethodCall:
		return !astquery.Equivalent(ctx.File, c.Var, recv)
	case *syntax.FuncCall:
		return semanticquery.NativeBuiltinName(ctx, c) == ""
	case *syntax.StaticCall:
		return !stubClass(ctx, c.Class)
	case *syntax.New:
		return !stubClass(ctx, c.Class)
	case *syntax.Include:
		return true
	}
	return false
}

func stubClass(ctx *analysis.Context, e syntax.Expr) bool {
	name, ok := e.(*syntax.Name)
	if !ok {
		return false
	}
	c := ctx.Index().Class(ctx.Names().Class(name.Value, name.Span().Start), ctx.PHP)
	return c != nil && strings.HasPrefix(c.File, "stubs/")
}
