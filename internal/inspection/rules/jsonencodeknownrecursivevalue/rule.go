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
func cyclic(ctx *analysis.Context, at syntax.Node, root string) bool {
	arrays := map[string]bool{}
	edges := map[string]map[string]string{}
	for _, event := range semanticquery.ExpansionDLocalEvents(ctx, at) {
		if event.Span().Start >= at.Span().Start {
			break
		}
		if a, ok := event.(*syntax.Assign); ok {
			if v, ok := a.Var.(*syntax.Variable); ok {
				if !semanticquery.NativeDominates(a, at) {
					return false
				}
				_, array := syntax.UnwrapParens(a.Value).(*syntax.Array)
				arrays[v.Name] = array && !a.ByRef && a.Op.Kind == syntax.TEqual
				delete(edges, v.Name)
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
				delete(edges[v.Name], key)
				if to, ref := syntax.UnwrapParens(a.Value).(*syntax.Variable); ref && a.ByRef && a.Op.Kind == syntax.TEqual && arrays[v.Name] && arrays[to.Name] {
					edges[v.Name][key] = to.Name
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
	return cycle(edges, root, map[string]bool{}, map[string]bool{})
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
