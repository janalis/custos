package flow

import (
	"maps"
	"strconv"
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

// Resolve only audited constants used by finite collection instructions.
// Resolve imports and namespace declarations before accepting the native name.
func (c *pureCompiler) nativeConstant(node *syntax.ConstFetch) *pureExpr {
	name, fallback := c.env.types.Names.Const(node.Name.Value, node.Span().Start)
	for _, candidate := range []string{name, fallback} {
		if candidate == "" {
			continue
		}
		declaration := c.env.types.Index.Constant(candidate, c.env.types.PHP)
		if declaration == nil {
			continue
		}
		value, accepted := map[string]string{"COUNT_NORMAL": "0", "ARRAY_FILTER_USE_BOTH": "1", "ARRAY_FILTER_USE_KEY": "2"}[candidate]
		if !accepted || candidate != "COUNT_NORMAL" && c.env.types.PHP.Below(phpversion.PHP56) || !declaration.Builtin || !declaration.Avail.In(c.env.types.PHP) || strings.TrimSpace(declaration.Value) != value {
			return nil
		}
		number, _ := strconv.ParseInt(value, 10, 64)
		return pureConstant(constantScalar{k: 'i', i: number})
	}
	return nil
}

func (c *pureCompiler) stringPredicate(node *syntax.FuncCall, name string, params []index.Param) *pureExpr {
	if c.env.types.PHP.Below(phpversion.PHP80) || len(params) != 2 {
		return nil
	}
	args := c.arguments(node, []pureParam{{name: params[0].Name}, {name: params[1].Name}}, params)
	if args == nil {
		return nil
	}
	result := pureNode(pureStringPredicate, args[0], args[1], nil)
	if result != nil {
		result.projects = true
		result.slot = 0
		if name == "str_ends_with" {
			result.slot = 1
		}
		if name == "str_contains" {
			result.slot = 2
		}
	}
	return result
}

func (e *Env) evalStringPredicate(expr *pureExpr, left Value, slots []Value, r *scopeResult, depth int) (Value, bool) {
	right, ok := e.evalPure(expr.right, slots, r, depth-1)
	a, ak := knownScalar(left)
	b, bk := knownScalar(right)
	if !ok || !ak || !bk || a.k != 's' || b.k != 's' || len(a.s)+len(b.s) > 4096 || len(a.s)+len(b.s) > MaxTransfers-r.operations {
		return Value{}, false
	}
	r.operations += len(a.s) + len(b.s)
	found := strings.HasPrefix(a.s, b.s)
	if expr.slot == 1 {
		found = strings.HasSuffix(a.s, b.s)
	}
	if expr.slot == 2 {
		found = strings.Contains(a.s, b.s)
	}
	return scalarValue(constantScalar{k: 'b', b: found}), true
}

func (c *pureCompiler) spreadArray(array *syntax.Array) *pureExpr {
	if c.env.types.PHP.Below(phpversion.PHP74) {
		return nil
	}
	return c.runtimeArray(array)
}

// array checks cardinality and short-array syntax before either constructor.
func (c *pureCompiler) runtimeArray(array *syntax.Array) *pureExpr {
	expr := &pureExpr{kind: pureSpreadArray, cost: 1, depth: 1, projects: true, weakArguments: c.env.types.PHP.Below(phpversion.PHP81)}
	for _, item := range array.Items {
		if item == nil || item.ByRef || item.Unpack && c.env.types.PHP.Below(phpversion.PHP74) {
			return nil
		}
		var key *pureExpr
		if !item.Unpack {
			key = pureConstant(constantScalar{k: 'n'})
			key.appendKey = true
			if item.Key != nil {
				fuel := 256
				scalar, known := constantValue(item.Key, MaxDepth, &fuel)
				text, valid := arrayKey(item.Key)
				if !known {
					key = c.expression(item.Key)
					if key == nil {
						return nil
					}
				} else {
					if !valid || scalar.k == 'f' {
						return nil
					}
					// Exclude version-sensitive negative implicit indexing and overflow.
					if number, err := strconv.ParseInt(text, 10, 64); err == nil && strconv.FormatInt(number, 10) == text && (number < 0 || number == 9223372036854775807) {
						return nil
					}
					key = pureConstant(constantScalar{k: 's', s: text})
				}
			}
		}
		value := c.expression(item.Value)
		if value == nil {
			return nil
		}
		expr.args = append(expr.args, key, value)
		keyCost, keyDepth := 1, 1
		if key != nil {
			keyCost, keyDepth = key.cost, key.depth
		}
		expr.cost += keyCost + value.cost
		expr.depth = max(expr.depth, max(value.depth, keyDepth)+1)
	}
	if expr.cost > pureLimit || expr.depth > MaxDepth {
		return nil
	}
	return expr
}

// Each published intermediate stays immutable, including a carry read by the
// next iteration's projection. Bound both copies and the resulting shape.
func appendPureEntry(value *Value, entry Value, r *scopeResult) bool {
	set := value.elements
	r.copies += len(set.items) + 1
	if r.copies > MaxCopied || len(set.items) >= pureLimit {
		return false
	}
	copy := *set
	copy.items, copy.order = maps.Clone(set.items), append([]string(nil), set.order...)
	key := strconv.Itoa(len(set.items))
	copy.width++
	copy.put(key, scalarSnapshot(entry))
	copy.order = append(copy.order, key)
	*value = merge(*value, entry)
	value.elements, value.plain, value.deep = &copy, true, true
	return true
}

func (e *Env) evalSpreadArray(expr *pureExpr, slots []Value, r *scopeResult, depth int) (Value, bool) {
	set := &elementSet{whole: true, items: map[string]Value{}}
	result := Value{Complete: true, plain: true, deep: true, elements: set}
	next := int64(0)
	for i := 0; i < len(expr.args); i += 2 {
		if next > 9223372036854775807-pureLimit {
			return Value{}, false
		}
		var keyValue Value
		if expr.args[i] != nil {
			value, ok := e.evalPure(expr.args[i], slots, r, depth-1)
			if !ok {
				return Value{}, false
			}
			keyValue = value
		}
		entry, ok := e.evalPure(expr.args[i+1], slots, r, depth-1)
		if !ok || !pureReadOnlyInputs([]Value{entry}, nil) {
			return Value{}, false
		}
		keys, values := []string{}, []Value{}
		if expr.args[i] == nil {
			if !purePlainRecord(entry) || len(entry.elements.order) != len(entry.elements.items) || len(entry.elements.items) > pureLimit {
				return Value{}, false
			}
			for _, key := range entry.elements.order {
				if number, err := strconv.ParseInt(key, 10, 64); err == nil && strconv.FormatInt(number, 10) == key {
					key = strconv.FormatInt(next, 10)
					next++
				} else if expr.weakArguments {
					return Value{}, false
				}
				keys = append(keys, key)
			}
			for _, key := range entry.elements.order {
				values = append(values, entry.elements.items[key])
			}
		} else {
			key, scalar := knownScalar(keyValue)
			text := ""
			if scalar && key.k == 'n' && expr.args[i].appendKey {
				text = strconv.FormatInt(next, 10)
				next++
			} else {
				var valid bool
				text, valid = pureRecordKey(keyValue)
				if !valid {
					return Value{}, false
				}
				if number, err := strconv.ParseInt(text, 10, 64); err == nil && strconv.FormatInt(number, 10) == text {
					if number < 0 || number == 9223372036854775807 {
						return Value{}, false
					}
					next = max(next, number+1)
				}
			}
			keys, values = append(keys, text), append(values, entry)
		}
		r.copies += len(keys)
		r.operations += len(keys)
		if r.copies > MaxCopied || r.operations > MaxTransfers || len(set.items)+len(keys) > pureLimit {
			return Value{}, false
		}
		for j, key := range keys {
			if _, exists := set.items[key]; !exists {
				set.order = append(set.order, key)
				set.width++
			}
			set.put(key, scalarSnapshot(values[j]))
			result = merge(result, values[j])
		}
	}
	result.elements, result.plain, result.deep = set, true, true
	return result, true
}
