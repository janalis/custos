// Package jsonencodeknownrecursivevalue implements the native JsonEncodeKnownRecursiveValue inspection.
package jsonencodeknownrecursivevalue

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Remove the cycle before JSON encoding."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "JsonEncodeKnownRecursiveValue" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "json_encode") {
		return
	}
	flags := semanticquery.CallArgument(c.Args, 1, "flags")
	if flags != nil {
		bits, k := semanticquery.NativeContractInt(ctx, flags)
		if !k || bits&512 != 0 {
			return
		}
	}
	v, ok := syntax.UnwrapParens(semanticquery.CallArgument(c.Args, 0, "value")).(*syntax.Variable)
	if !ok {
		return
	}
	if cyclic(ctx, c, v.Name) {
		ctx.ReportNode(c, message)
	}
}

// cyclic follows explicit, straight-line local array-reference edges only.
// Every other use of a variable taking part in the graph removes the proof.
func cyclic(ctx *analysis.Context, at syntax.Node, root string) bool {
	arrays := map[string]bool{}
	edges := map[string]map[string]string{}
	modelled := map[syntax.Node]bool{}
	for _, event := range semanticquery.ExpansionDLocalEvents(ctx, at) {
		if event.Span().Start >= at.Span().Start {
			break
		}
		// An expression enclosing the call completes only after it.
		if event.Span().End > at.Span().Start {
			continue
		}
		if a, ok := event.(*syntax.Assign); ok {
			if v, ok := a.Var.(*syntax.Variable); ok {
				if !semanticquery.NativeDominates(a, at) {
					return false
				}
				_, array := syntax.UnwrapParens(a.Value).(*syntax.Array)
				arrays[v.Name] = array && !a.ByRef && a.Op.Kind == syntax.TEqual
				delete(edges, v.Name)
				modelled[v] = true
			}
			if slot, ok := a.Var.(*syntax.ArrayDimFetch); ok {
				v, local := slot.Var.(*syntax.Variable)
				key, k := semanticquery.NativeArrayKey(ctx, slot.Dim)
				if !local || !k || !semanticquery.NativeDominates(a, at) {
					return false
				}
				if edges[v.Name] == nil {
					edges[v.Name] = map[string]string{}
				}
				// A plain write to a reference slot also replaces the referenced value.
				if _, bound := edges[v.Name][key]; bound && !a.ByRef {
					return false
				}
				delete(edges[v.Name], key)
				modelled[v] = true
				if to, ref := syntax.UnwrapParens(a.Value).(*syntax.Variable); ref && a.ByRef && a.Op.Kind == syntax.TEqual && arrays[v.Name] && arrays[to.Name] {
					edges[v.Name][key] = to.Name
					modelled[to] = true
				}
			}
		}
		if _, ok := event.(*syntax.Unset); ok {
			return false
		}
		if _, ok := event.(*syntax.FuncCall); ok {
			return false
		}
		if _, ok := event.(*syntax.MethodCall); ok {
			return false
		}
	}
	return cycle(edges, root, map[string]bool{}, map[string]bool{}) && undisturbed(ctx, at, edges, modelled)
}

// undisturbed rejects every construct before at that may change a graph
// variable without being one of the modelled assignments: destructuring and
// foreach targets, increments, aliases, copies, dynamic variables, included
// code and the calls the event list does not carry.
func undisturbed(ctx *analysis.Context, at syntax.Node, edges map[string]map[string]string, modelled map[syntax.Node]bool) bool {
	tracked := map[string]bool{}
	for from, slots := range edges {
		tracked[from] = true
		for _, to := range slots {
			tracked[to] = true
		}
	}
	scope := syntax.EnclosingVariableScope(at)
	ok := true
	visit := func(n syntax.Node) bool {
		if !ok || n.Span().Start >= at.Span().Start || n != scope && syntax.IsVariableScope(n) {
			return false
		}
		switch n := n.(type) {
		case *syntax.Variable:
			ok = n.NameExpr == nil && (!tracked[n.Name] || modelled[n])
		case *syntax.StaticCall, *syntax.New, *syntax.Include, *syntax.Eval:
			ok = false
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

func cycle(edges map[string]map[string]string, name string, active, done map[string]bool) bool {
	if active[name] {
		return true
	}
	if done[name] {
		return false
	}
	active[name] = true
	for _, next := range edges[name] {
		if cycle(edges, next, active, done) {
			return true
		}
	}
	delete(active, name)
	done[name] = true
	return false
}
