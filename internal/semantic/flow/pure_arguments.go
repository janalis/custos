package flow

import (
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

func (c *pureCompiler) native(node *syntax.FuncCall, name string, params []index.Param) *pureExpr {
	if name == "str_replace" || name == "str_repeat" {
		return c.byteTransform(node, name, params)
	}
	switch name {
	case "array_fill", "array_fill_keys", "array_flip", "array_chunk", "array_sum":
		return c.arrayConstruction(node, name, params)
	case "array_pad", "range":
		return c.padding(node, name, params)
	case "array_unique", "array_search", "array_product", "strrev", "substr_count":
		return c.query(node, name, params)
	case "min", "max":
		return c.extreme(node, name)
	case "strtolower", "strtoupper", "ucfirst", "lcfirst", "ltrim", "rtrim", "str_pad":
		return c.textShape(node, name, params)
	case "trim":
		if node.Args != nil && len(node.Args.Args) == 2 {
			return c.textShape(node, name, params)
		}
	}
	if name == "array_reverse" {
		return c.reverse(node, params)
	}
	if name == "implode" || name == "explode" {
		return c.stringTransform(node, name, params)
	}
	if name == "array_column" {
		return c.column(node, params)
	}
	if name == "array_combine" || name == "array_replace" {
		return c.pairedArray(node, name, params)
	}
	if name == "array_merge" {
		return c.arrayMerge(node, params)
	}
	if name == "array_reduce" {
		return c.fold(node, params)
	}
	if name == "array_slice" {
		return c.slice(node, params)
	}
	if name == "array_map" || name == "array_filter" {
		return c.higher(node, name, params)
	}
	switch name {
	case "array_values", "array_keys", "array_key_first", "array_key_last", "is_string", "is_int", "is_array", "substr":
		return c.collectionNative(node, name, params)
	}

	if name == "str_starts_with" || name == "str_ends_with" || name == "str_contains" {
		return c.stringPredicate(node, name, params)
	}
	if name == "count" || name == "in_array" {
		return c.finiteNative(node, name, params)
	}
	if name == "array_key_exists" {
		return c.keyExists(node, params)
	}
	if name != "strlen" && name != "trim" || node.Args == nil || len(node.Args.Args) != 1 || len(params) == 0 {
		return nil
	}
	arg, ok := node.Args.Args[0].(*syntax.Arg)
	if !ok || arg.ByRef || arg.Unpack || arg.Name != nil && (c.env.types.PHP.Below(phpversion.PHP80) || arg.Name.Value != params[0].Name) {
		return nil
	}
	child := c.expression(arg.Value)
	if child == nil || child.cost >= pureLimit || child.depth >= MaxDepth {
		return nil
	}
	kind := pureLength
	if name == "trim" {
		kind = pureTrim
	}
	return &pureExpr{kind: kind, left: child, projects: child.projects, cost: child.cost + 1, depth: child.depth + 1}
}

// Ordinary calls keep their compact compiler binding; spreads use the same
// runtime role binder as value analysis, with scalar indices for symbolic terms.
func (c *pureCompiler) arguments(node syntax.Node, params []pureParam, declared []index.Param) []*pureExpr {
	// Every composed call requires the published contract and declaration to agree.
	// An editor can replace declaration metadata before an old snapshot expires.
	if len(declared) != len(params) {
		return nil
	}
	for i, param := range declared {
		if param.ByRef || param.Variadic || param.Name != params[i].name {
			return nil
		}
	}
	if arguments(node) != nil {
		for _, item := range arguments(node).Args {
			if arg, ok := item.(*syntax.Arg); ok && arg.Unpack {
				return c.spreadArguments(node, params, declared)
			}
		}
	}
	args := make([]*pureExpr, len(params))
	position := 0
	named := false
	if arguments(node) != nil {
		for _, item := range arguments(node).Args {
			arg, ok := item.(*syntax.Arg)
			if !ok || arg.ByRef {
				return nil
			}
			at := position
			if arg.Name != nil {
				if c.env.types.PHP.Below(phpversion.PHP80) {
					return nil
				}
				named, at = true, -1
				for i, param := range params {
					if param.name == arg.Name.Value {
						at = i
						break
					}
				}
			} else {
				if named {
					return nil
				}
				position++
			}
			if at < 0 || at >= len(args) || args[at] != nil {
				return nil
			}
			args[at] = c.expression(arg.Value)
			if args[at] == nil {
				return nil
			}
		}
	}
	return args
}

func (c *pureCompiler) spreadArguments(node syntax.Node, params []pureParam, declared []index.Param) []*pureExpr {
	var terms []*pureExpr
	marker := func(term *pureExpr) Value {
		index := len(terms)
		terms = append(terms, term)
		return scalarValue(constantScalar{k: 'i', i: int64(index)})
	}
	values := make([]Value, len(arguments(node).Args))
	for i, item := range arguments(node).Args {
		arg, ok := item.(*syntax.Arg)
		if !ok || arg.ByRef {
			return nil
		}
		term := c.expression(arg.Value)
		if term == nil {
			return nil
		}
		if !arg.Unpack {
			values[i] = marker(term)
			continue
		}
		if term.kind != pureArray {
			return nil
		}
		set := &elementSet{whole: true, items: map[string]Value{}, width: len(term.args) / 2}
		for j := 0; j < len(term.args); j += 2 {
			key := term.args[j].literal.s
			set.put(key, marker(term.args[j+1]))
			set.order = append(set.order, key)
		}
		values[i] = Value{Complete: true, elements: set, plain: true, deep: true}
	}
	bound, seen, known := c.env.bindValues(arguments(node), declared, values, false)
	if !known {
		return nil
	}
	args := make([]*pureExpr, len(params))
	for i, given := range seen {
		if given {
			args[i] = terms[bound[i].literal.i]
		}
	}
	return args
}
