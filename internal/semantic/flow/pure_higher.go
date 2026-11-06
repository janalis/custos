package flow

import (
	"slices"
	"strconv"
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
	// The third map role is a single bounded positional array, not an
	// unbounded variadic pack. Keep the original metadata immutable.
	declared := slices.Clone(params)
	if name == "array_map" {
		declared[2].Variadic = false
	}
	roles := make([]pureParam, 3)
	for i, param := range declared {
		roles[i].name = param.Name
	}
	previousInline := c.inlineCallbacks
	c.inlineCallbacks = true
	args := c.arguments(node, roles, declared)
	c.inlineCallbacks = previousInline
	if args == nil || args[0] == nil || args[1] == nil {
		return nil
	}
	callback, input := args[0], args[1]
	mode, arity := 0, 1
	var second *pureExpr
	if name == "array_filter" {
		input, callback = args[0], args[1]
		mode = 1 // value callback
		if args[2] != nil {
			if c.env.types.PHP.Below(phpversion.PHP56) || args[2].kind != pureLiteral || args[2].literal.k != 'i' {
				return nil
			}
			switch args[2].literal.i {
			case 0:
			case 1:
				mode, arity = 3, 2 // value, key
			case 2:
				mode = 2 // key
			default:
				return nil
			}
		}
	} else if args[2] != nil {
		// PHP does not accept named arguments for the variadic arrays role.
		// Keep this initial two-input form strictly positional, including packs.
		for _, item := range node.Args.Args {
			arg, ok := item.(*syntax.Arg)
			if !ok || arg.Name != nil || arg.Unpack {
				return nil
			}
		}
		second, arity = args[2], 2
	}
	var invoke *pureExpr
	if callback.kind == pureInline {
		invoke = c.callbackInvoke(callback.call, arity)
	} else if callback.kind == pureLiteral && callback.literal.k == 's' {
		key := callbackName(callback.literal.s)
		function := c.env.types.Index.Function(key, c.env.types.PHP)
		if function == nil || function.ByRef || !function.Avail.In(c.env.types.PHP) {
			return nil
		}
		if function.Builtin {
			if name != "array_map" || second != nil || key != "strlen" && key != "trim" {
				return nil
			}
			invoke = &pureExpr{kind: pureLength, left: &pureSlots[0], projects: true, cost: 2, depth: 2}
			if key == "trim" {
				invoke.kind = pureTrim
			}
		} else {
			invoke = c.namedCallback(key, arity)
		}
	}
	if invoke == nil {
		return nil
	}
	result := pureNode(pureHigher, input, invoke, second)
	if result == nil {
		return nil
	}
	result.projects, result.slot = true, mode
	return result
}

func (e *Env) evalPureHigher(expr *pureExpr, input Value, slots []Value, r *scopeResult, depth int) (Value, bool) {
	if !purePlainRecord(input) || len(input.elements.order) != len(input.elements.items) {
		return Value{}, false
	}
	var second Value
	if expr.other != nil {
		var ok bool
		second, ok = e.evalPure(expr.other, slots, r, depth-1)
		if !ok || !purePlainRecord(second) || len(second.elements.order) != len(second.elements.items) || len(second.elements.order) != len(input.elements.order) {
			return Value{}, false
		}
	}
	size := len(input.elements.items)
	r.copies += size
	if size > pureLimit || r.copies > MaxCopied {
		return Value{}, false
	}
	set := &elementSet{whole: true, items: map[string]Value{}, width: size}
	result := Value{Complete: true, plain: true, deep: true}
	for i, key := range input.elements.order {
		original := input.elements.items[key]
		args := []Value{original}
		if expr.slot == 2 {
			args[0] = scalarValue(pureKeyScalar(key))
		}
		if expr.slot == 3 {
			args = append(args, scalarValue(pureKeyScalar(key)))
		}
		if expr.other != nil {
			args = append(args, second.elements.items[second.elements.order[i]])
		}
		value, ok := e.evalPure(expr.right, args, r, depth-1)
		if !ok || !pureReadOnlyInputs([]Value{value}, nil) {
			return Value{}, false
		}
		if expr.slot != 0 {
			keep, known := pureTruthValue(value)
			if !known {
				return Value{}, false
			}
			if !keep {
				continue
			}
			value = original
		}
		if expr.other != nil {
			key = strconv.Itoa(i)
		}
		set.put(key, scalarSnapshot(value))
		set.order = append(set.order, key)
		result = merge(result, value)
	}
	result.elements, result.plain, result.deep = set, true, true
	return result, true
}

func (c *pureCompiler) namedCallback(key string, arity int) *pureExpr {
	function := c.env.types.Index.Function(key, c.env.types.PHP)
	if function == nil || function.Builtin || function.ByRef || !function.Avail.In(c.env.types.PHP) {
		return nil
	}
	if len(c.env.types.Index.FunctionDecls(key, c.env.types.PHP)) != 1 {
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
	if callee.outer != 0 || len(callee.params) < arity || len(callee.params) != len(function.Params) || !callee.result.Equal(types.FromDoc(function.Return, nil)) || key == c.name || slices.Contains(callee.calls, c.name) {
		return nil
	}

	for i, param := range callee.params {
		declared := function.Params[i]
		if c.env.types.PHP.Below(phpversion.PHP70) && declared.Type != "" && declared.Type != "array" {
			return nil
		}
		if declared.ByRef || declared.Variadic || declared.Name != param.name || !param.typ.Equal(types.FromDoc(declared.Type, nil)) || param.optional != declared.Optional {
			return nil
		}
	}
	invoke := c.callbackInvoke(callee, arity)
	c.calls = append(c.calls, key)
	return invoke
}

func (c *pureCompiler) callbackInvoke(callee *pureContract, arity int) *pureExpr {
	if len(callee.params) < arity || callee.outer != 0 || slices.Contains(callee.calls, c.name) {
		return nil
	}
	invoke := &pureExpr{kind: pureCall, call: callee, weakArguments: true, projects: true, cost: callee.expr.cost + 1, depth: callee.expr.depth + 1}
	for i, param := range callee.params {
		value := &pureSlots[i]
		if i >= arity {
			if !param.optional || param.initial == nil {
				return nil
			}
			value = param.initial
		}
		invoke.args = append(invoke.args, value)
		invoke.cost += value.cost
		invoke.depth = max(invoke.depth, value.depth+1)
	}
	c.calls = append(c.calls, callee.calls...)
	return invoke
}

func callbackName(name string) string { return strings.ToLower(strings.TrimPrefix(name, `\`)) }
