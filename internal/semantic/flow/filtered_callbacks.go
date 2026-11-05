package flow

import (
	"strconv"
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// A finite filter invokes callbacks for sinks, but never replaces an entry with
// the predicate's result. Unknown predicates retain possible original entries.
func (e *Env) filteredCallback(at syntax.Node, list *syntax.ArgList, values []Value, f *frame, r *scopeResult) (Value, bool) {
	call, ok := at.(*syntax.FuncCall)
	if !ok {
		return Value{}, false
	}
	bound, seen, known := e.callbackBindings(at, list, values, false)
	if !known || len(bound) != 3 {
		return Value{}, false
	}
	input, callee := bound[0], bound[1]
	if !input.Complete || input.replaced || input.Invalidated != 0 || input.elements == nil || !input.elements.whole {
		return Value{}, false
	}
	nullCallback := !seen[1] || constantName(callee.Expr) == "null"
	if !nullCallback && (!callee.Complete || callee.replaced || callee.Invalidated != 0) {
		return Value{}, false
	}
	mode := int64(0)
	if seen[2] && e.types.PHP.Below(phpversion.PHP56) {
		return Value{}, false
	}
	if !nullCallback && seen[2] {
		var valid bool
		mode, valid = e.htmlFlagBits(bound[2].Expr, r, 16)
		if !valid || mode < 0 || mode > 2 {
			return Value{}, false
		}
	}
	keys := input.elements.order
	if len(keys) != len(input.elements.items) {
		return Value{}, false
	}
	if len(keys) == 0 {
		return Value{Complete: true, Expr: call, plain: true, deep: true, elements: emptyElements}, true
	}
	prior, had := r.closureSinks[at]
	selected := &elementSet{items: map[string]Value{}, width: len(keys), whole: true, optional: input.elements.optional}
	result := Value{Complete: true, Expr: call, plain: true, deep: true}
	for _, key := range keys {
		original := input.elements.items[key]
		predicate := original
		if !nullCallback {
			arguments := []Value{original}
			keyValue := constantArrayKeyValue(key)
			switch mode {
			case 1:
				arguments = append(arguments, keyValue)
			case 2:
				arguments = []Value{keyValue}
			}
			forwarded := &syntax.ArgList{}
			for _, value := range arguments {
				expr := value.Expr
				if expr == nil {
					expr = call
				}
				forwarded.Args = append(forwarded.Args, &syntax.Arg{Value: expr})
			}
			r.copies += len(f.vars) + len(f.states)
			branch := clone(f)
			var valid bool
			if callee.returned != nil {
				predicate, _, valid = e.invokeReturned(at, callee.returned, forwarded, arguments, r)
			} else {
				predicate, _, valid = e.invokeCallableSet(at, callee, forwarded, arguments, &branch, r)
			}
			if !valid {
				if had {
					r.closureSinks[at] = prior
				} else {
					delete(r.closureSinks, at)
				}
				return Value{}, false
			}
		}
		truth, certain := constantTruth(predicate.Expr)
		if !predicate.Complete {
			certain = false
		}
		if certain && !truth {
			continue
		}
		selected.optional = selected.optional || !certain
		selected.put(key, elementValue(original))
		selected.order = append(selected.order, key)
		result = merge(result, original)
	}
	if !input.deep {
		return Value{}, false
	}
	if len(selected.items) == 0 {
		selected = emptyElements
	}
	result.Expr, result.elements, result.plain, result.deep = call, selected, true, true
	return result, true
}

func constantArrayKeyValue(key string) Value {
	raw, kind := key, syntax.LitInt
	if number, err := strconv.ParseInt(key, 10, 64); err != nil || strconv.FormatInt(number, 10) != key {
		kind = syntax.LitString
		raw = "'" + strings.ReplaceAll(strings.ReplaceAll(key, `\`, `\\`), `'`, `\'`) + "'"
	}
	return Value{Complete: true, Expr: &syntax.Literal{LitKind: kind, Raw: raw}, plain: true, deep: true}
}
