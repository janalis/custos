package infer

import (
	"custos/internal/syntax"
	"custos/internal/types"
)

// promotedHookDocSpans lists hook ranges nested in a scope's parameters in
// source order. Their inline annotations belong to the hooks, not the scope.
// Ordinary parameters leave the nil result unallocated.
func promotedHookDocSpans(scope syntax.Node) []syntax.Span {
	var spans []syntax.Span
	for _, p := range syntax.VariableScopeParams(scope) {
		for _, h := range p.Hooks {
			spans = append(spans, h.Span())
		}
	}
	return spans
}

// hookValueType is the default setter parameter type. A promoted property's
// declaration is its constructor parameter, but its hook has separate locals.
func (e *Env) hookValueType(h *syntax.PropertyHook) types.Type {
	var t types.Type
	switch p := h.Parent().(type) {
	case *syntax.Property:
		at := p.Span().Start
		t = types.FromNode(p.Type, e.resolver(at))
		if !e.native {
			if d := e.DocOf(p); d != nil {
				if doc := d.VarType(""); doc != "" {
					t = pickMemberType(t, types.FromDoc(doc, e.resolverFor(p, at)))
				}
			}
		}
	case *syntax.Param:
		t = e.paramType(syntax.EnclosingFuncLike(p), p)
	}
	cls := e.selfClass(syntax.EnclosingClass(h))
	if t.Has("parent") {
		c := e.Index.Class(cls, e.PHP)
		if c == nil || c.Parent == "" {
			return types.Unknown
		}
		parent := types.Of(`\` + c.Parent)
		if rest := t.Without("parent"); !rest.IsUnknown() {
			parent = types.Union(rest, parent)
		}
		t = parent
	}
	return bindStatic(t, cls)
}
