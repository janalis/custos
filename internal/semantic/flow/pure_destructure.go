package flow

import (
	"strconv"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

type pureField struct{ name, key string }

// Only one-level plain variable targets can enter an immutable loop program.
func (c *pureCompiler) loopFields(value syntax.Expr, carry string) ([]pureField, bool) {
	return c.bindingFields(value, carry, false)
}

func (c *pureCompiler) bindingFields(value syntax.Expr, carry string, local bool) ([]pureField, bool) {
	if name := plainVariable(value); name != "" {
		if name == carry || external(name) || name == "GLOBALS" || name == "this" {
			return nil, false
		}
		return []pureField{{name: name}}, false
	}
	var items []*syntax.ArrayItem
	switch node := value.(type) {
	case *syntax.List:
		if !local && c.env.types.PHP.Below(phpversion.PHP55) {
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

func (c *pureCompiler) destructure(assign *syntax.Assign, tail []syntax.Stmt, depth int) *pureExpr {
	fields, _ := c.bindingFields(assign.Var, "", true)
	if fields == nil {
		return nil
	}
	value := c.expression(assign.Value)
	if value == nil {
		return nil
	}
	slot := c.nextSlot
	c.nextSlot = slot + 3 + len(fields)
	for i, field := range fields {
		c.bind(field.name, &pureExpr{kind: pureSlot, slot: slot + 3 + i, cost: 1, depth: 1})
	}
	result := pureNode(pureDestructure, value, c.statements(tail, depth-1), nil)
	if result == nil {
		return nil
	}
	result.cost += len(fields)
	if result.cost > pureLimit {
		return nil
	}
	result.slot, result.destruct, result.projects = slot, fields, true
	return result
}

func (e *Env) evalDestructure(expr *pureExpr, slots []Value, r *scopeResult, depth int) (Value, bool) {
	value, ok := e.evalPure(expr.left, slots, r, depth-1)
	if !ok {
		return Value{}, false
	}
	size := max(len(slots), expr.slot+3+len(expr.destruct))
	r.copies += size
	if r.copies > MaxCopied {
		return Value{}, false
	}
	bound := make([]Value, size)
	copy(bound, slots)
	if !bindPureFields(expr.destruct, value, bound, expr.slot, r) {
		return Value{}, false
	}
	return e.evalPure(expr.right, bound, r, depth-1)
}
