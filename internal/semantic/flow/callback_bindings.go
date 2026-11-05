package flow

import "custos/internal/php/syntax"

// Callback bindings use the shared runtime-role binder. An expanded pack must
// also retain every captured entry; a sealed type alone cannot prove that.
func (e *Env) callbackBindings(at syntax.Node, list *syntax.ArgList, values []Value, packs bool) ([]Value, []bool, bool) {
	if list == nil || len(list.Args) != len(values) {
		return nil, nil, false
	}
	for i, node := range list.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok || arg == nil || arg.ByRef || arg.Unpack && (values[i].elements == nil || !values[i].elements.whole || values[i].elements.optional) {
			return nil, nil, false
		}
	}
	params, _, resolved := e.declaration(at)
	if !resolved {
		return nil, nil, false
	}
	return e.bindValues(list, params, values, packs)
}
