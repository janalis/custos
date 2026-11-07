package flow

import (
	"math"

	"custos/internal/php/syntax"
)

func (c *pureCompiler) integer(op syntax.TokenKind, left, right syntax.Expr) *pureExpr {
	a := c.expression(left)
	if a == nil {
		return nil
	}
	if op == syntax.TTilde {
		if a.cost >= pureLimit || a.depth >= MaxDepth {
			return nil
		}
		return &pureExpr{kind: pureInteger, left: a, op: op, cost: a.cost + 1, depth: a.depth + 1, projects: true}
	}
	result := pureNode(pureInteger, a, c.expression(right), nil)
	if result != nil {
		result.op, result.projects = op, true
	}
	return result
}

func (e *Env) evalPureInteger(expr *pureExpr, left Value, slots []Value, r *scopeResult, depth int) (Value, bool) {
	a, known := knownScalar(left)
	if !known || a.k != 'i' || !pureReadOnlyInputs([]Value{left}, nil) || len(left.Sources) > MaxSources {
		return Value{}, false
	}
	sources := left.Sources
	var number int64
	if expr.op == syntax.TTilde {
		number = ^a.i
	} else {
		right, ok := e.evalPure(expr.right, slots, r, depth-1)
		b, bk := knownScalar(right)
		if !ok || !bk || b.k != 'i' || !pureReadOnlyInputs([]Value{right}, nil) || len(right.Sources) > MaxSources {
			return Value{}, false
		}
		var complete bool
		sources, complete = union(sources, right.Sources)
		if !complete {
			return Value{}, false
		}
		switch expr.op {
		case syntax.TDiv:
			if b.i == 0 || a.i == math.MinInt64 && b.i == -1 || a.i%b.i != 0 {
				return Value{}, false
			}
			number = a.i / b.i
		case syntax.TMod:
			if b.i == 0 {
				return Value{}, false
			}
			if a.i == math.MinInt64 && b.i == -1 {
				number = 0
			} else {
				number = a.i % b.i
			}
		case syntax.TAmpersand:
			number = a.i & b.i
		case syntax.TBar:
			number = a.i | b.i
		case syntax.TCaret:
			number = a.i ^ b.i
		case syntax.TSl, syntax.TSr:
			if b.i < 0 || b.i > 63 {
				return Value{}, false
			}
			if expr.op == syntax.TSl {
				if a.i > math.MaxInt64>>uint(b.i) || a.i < math.MinInt64>>uint(b.i) {
					return Value{}, false
				}
				number = a.i << uint(b.i)
			} else {
				number = a.i >> uint(b.i)
			}
		default:
			return Value{}, false
		}
	}
	value := scalarValue(constantScalar{k: 'i', i: number})
	value.Sources, value.Safe = sources, everyContext
	return value, true
}
