package flow

import (
	"slices"
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

func (c *pureCompiler) higher(node *syntax.FuncCall, name string, params []index.Param) *pureExpr {
	if len(params) != 3 {
		return nil
	}
	// Model only the first two runtime roles. Extra arrays or filter modes
	// must be rejected by the shared binder, including named/unpacked forms.
	roles := []pureParam{{name: params[0].Name}, {name: params[1].Name}}
	args := c.arguments(node, roles, params[:2])
	if args == nil || args[0] == nil || args[1] == nil {
		return nil
	}
	callback, input := args[0], args[1]
	if name == "array_filter" {
		input, callback = args[0], args[1]
	}
	if callback.kind != pureLiteral || callback.literal.k != 's' {
		return nil
	}
	key := strings.ToLower(strings.TrimPrefix(callback.literal.s, `\`))
	function := c.env.types.Index.Function(key, c.env.types.PHP)
	if function == nil || function.Builtin || function.ByRef || !function.Avail.In(c.env.types.PHP) || len(c.env.types.Index.FunctionDecls(key, c.env.types.PHP)) != 1 {
		return nil
	}
	summary, known := c.env.contract(key)
	if !known || !summary.Fixed || summary.Generator || summary.pure == nil {
		return nil
	}
	if c.env.types.PHP.Below(phpversion.PHP70) && function.Return != "" {
		return nil
	}
	callee := summary.pure
	if callee.outer != 0 || len(callee.params) == 0 || len(callee.params) != len(function.Params) || !callee.result.Equal(types.FromDoc(function.Return, nil)) || key == c.name || slices.Contains(callee.calls, c.name) {
		return nil
	}
	invoke := &pureExpr{kind: pureCall, call: callee, weakArguments: true, projects: true, cost: callee.expr.cost + 1, depth: callee.expr.depth + 1}
	for i, param := range callee.params {
		declared := function.Params[i]
		if c.env.types.PHP.Below(phpversion.PHP70) && declared.Type != "" && declared.Type != "array" {
			return nil
		}
		if declared.ByRef || declared.Variadic || declared.Name != param.name || !param.typ.Equal(types.FromDoc(declared.Type, nil)) || param.optional != declared.Optional {
			return nil
		}
		value := &pureSlots[0]
		if i != 0 {
			if !param.optional || param.initial == nil {
				return nil
			}
			value = param.initial
		}
		invoke.args = append(invoke.args, value)
		invoke.cost += value.cost
		invoke.depth = max(invoke.depth, value.depth+1)
	}
	result := pureNode(pureHigher, input, invoke, nil)
	if result == nil {
		return nil
	}
	result.projects = true
	if name == "array_filter" {
		result.slot = 1
	}
	c.calls = append(c.calls, key)
	c.calls = append(c.calls, callee.calls...)
	return result
}

func (e *Env) evalPureHigher(expr *pureExpr, input Value, slots []Value, r *scopeResult, depth int) (Value, bool) {
	if !purePlainRecord(input) || len(input.elements.order) != len(input.elements.items) {
		return Value{}, false
	}
	size := len(input.elements.items)
	r.copies += size
	if size > pureLimit || r.copies > MaxCopied {
		return Value{}, false
	}
	set := &elementSet{whole: true, items: map[string]Value{}, width: size}
	result := Value{Complete: true, plain: true, deep: true}
	for _, key := range input.elements.order {
		original := input.elements.items[key]
		value, ok := e.evalPure(expr.right, []Value{original}, r, depth-1)
		if !ok || !pureReadOnlyInputs([]Value{value}, nil) {
			return Value{}, false
		}
		if expr.slot == 1 {
			keep, known := pureTruthValue(value)
			if !known {
				return Value{}, false
			}
			if !keep {
				continue
			}
			value = original
		}
		set.put(key, scalarSnapshot(value))
		set.order = append(set.order, key)
		result = merge(result, value)
	}
	result.elements, result.plain, result.deep = set, true, true
	return result, true
}
