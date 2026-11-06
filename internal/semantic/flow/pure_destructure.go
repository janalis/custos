package flow

import (
	"strconv"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

type pureField struct{ name, key string }

// Only one-level plain variable targets can enter an immutable loop program.
func (c *pureCompiler) loopFields(value syntax.Expr, carry string) ([]pureField, bool) {
	if name := plainVariable(value); name != "" {
		if name == carry || external(name) || name == "GLOBALS" || name == "this" {
			return nil, false
		}
		return []pureField{{name: name}}, false
	}
	var items []*syntax.ArrayItem
	switch node := value.(type) {
	case *syntax.List:
		if c.env.types.PHP.Below(phpversion.PHP55) {
			return nil, false
		}
		items = node.Items
	case *syntax.Array:
		if !node.Short || c.env.types.PHP.Below(phpversion.PHP71) {
			return nil, false
		}
		items = node.Items
	default:
		return nil, false
	}
	if len(items) > pureLimit {
		return nil, false
	}
	fields := []pureField{}
	seen := map[string]bool{}
	keyed, positional := false, false
	for i, item := range items {
		if item != nil && (item.ByRef || item.Unpack || item.Value == nil && item.Key != nil) {
			return nil, false
		}
		if item == nil || item.Value == nil {
			positional = true
			continue
		}
		name := plainVariable(item.Value)
		if name == "" || name == carry || external(name) || name == "GLOBALS" || name == "this" || seen[name] {
			return nil, false
		}
		key := strconv.Itoa(i)
		if item.Key != nil {
			if c.env.types.PHP.Below(phpversion.PHP71) {
				return nil, false
			}
			fuel := 256
			scalar, known := constantValue(item.Key, MaxDepth, &fuel)
			if !known || scalar.k != 'i' && scalar.k != 's' {
				return nil, false
			}
			key = scalar.s
			if scalar.k == 'i' {
				key = strconv.FormatInt(scalar.i, 10)
			}
			keyed = true
		} else {
			positional = true
		}
		fields = append(fields, pureField{name: name, key: key})
		seen[name] = true
	}
	if len(fields) == 0 || keyed && positional {
		return nil, false
	}
	return fields, true
}

func bindPureFields(fields []pureField, item Value, bound []Value, slot int, r *scopeResult) bool {
	if !purePlainRecord(item) {
		return false
	}
	for i, field := range fields {
		r.operations++
		value, present := item.elements.lookup(field.key)
		if !present || r.operations > MaxTransfers || !pureReadOnlyInputs([]Value{value}, nil) {
			return false
		}
		bound[slot+3+i] = value
	}
	return true
}
