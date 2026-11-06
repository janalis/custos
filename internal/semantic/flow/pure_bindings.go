package flow

import "custos/internal/php/syntax"

func (e *Env) literalDefaults(scope syntax.Node) []Value {
	params := syntax.VariableScopeParams(scope)
	lastRequired := -1
	for i, p := range params {
		if p.Default == nil {
			lastRequired = i
		}
	}
	var defaults []Value
	for i, p := range params {
		if i <= lastRequired || p.Default == nil {
			continue
		}
		c := pureCompiler{env: e, defaults: true}
		expr := c.expression(p.Default)
		if expr == nil {
			continue
		}
		if value, known := e.evalPure(expr, nil, &scopeResult{}, MaxDepth); known {
			if defaults == nil {
				defaults = make([]Value, len(params))
			}
			defaults[i] = value
		}
	}
	return defaults
}

func (e *Env) defaultArguments(at syntax.Node, values, defaults []Value) ([]Value, bool) {
	params, _, known := e.declaration(at)
	if !known {
		return nil, false
	}
	actual, seen, known := e.bindValues(arguments(at), params, values, false)
	if !known {
		return nil, false
	}
	return callbackDefaults(actual, seen, defaults), true
}

func callbackDefaults(values []Value, seen []bool, defaults []Value) []Value {
	for i, given := range seen {
		if !given && i < len(defaults) && defaults[i].Complete {
			values[i] = defaults[i]
		}
	}
	return values
}

func pureCaptures(expr syntax.Expr, captures map[string]Value, outer int) map[string]*pureExpr {
	var slots map[string]*pureExpr
	params := syntax.VariableScopeParams(syntax.EnclosingVariableScope(expr))
	if len(params) != outer {
		return nil
	}
	for name, value := range captures {
		var projection *pureExpr
		if scalar, known := knownScalar(value); known {
			projection = pureConstant(scalar)
		} else if value.Complete && !value.replaced && value.Invalidated == 0 {
			for i, p := range params {
				if p.Var != nil && value.Identity == p.Var.Span().Start+1 {
					projection = &pureExpr{kind: pureSlot, slot: i, cost: 1, depth: 1}
					break
				}
			}
		}
		if projection != nil {
			if slots == nil {
				slots = map[string]*pureExpr{}
			}
			slots[name] = projection
		}
	}
	return slots
}
