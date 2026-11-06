package flow

import (
	"strconv"

	"custos/internal/php/syntax"
)

func (c *pureCompiler) cast(node *syntax.Unary) *pureExpr {
	value := c.expression(node.Expr)
	if value == nil || value.cost >= pureLimit || value.depth >= MaxDepth {
		return nil
	}
	return &pureExpr{kind: pureCast, left: value, op: node.Op.Kind, projects: true, cost: value.cost + 1, depth: value.depth + 1}
}

func evalPureCast(expr *pureExpr, input Value) (Value, bool) {
	if !pureReadOnlyInputs([]Value{input}, nil) {
		return Value{}, false
	}
	if expr.op == syntax.TBoolCast {
		truth, known := pureTruthValue(input)
		result := scalarValue(constantScalar{k: 'b', b: truth})
		result.Sources = input.Sources
		result.Safe = everyContext
		return result, known
	}
	scalar, known := knownScalar(input)
	if !known || scalar.k == 'f' {
		return Value{}, false
	}
	target := "int"
	if expr.op == syntax.TStringCast {
		target = "string"
	}
	var result constantScalar
	if scalar.k == 'n' {
		result = constantScalar{k: 'i'}
		if target == "string" {
			result = constantScalar{k: 's'}
		}
	} else if target == "int" && scalar.k == 'i' || target == "string" && scalar.k == 's' {
		result = scalar
	} else {
		if target == "int" && scalar.k == 's' {
			number, err := strconv.ParseInt(scalar.s, 10, 64)
			if err != nil || strconv.FormatInt(number, 10) != scalar.s {
				return Value{}, false
			}
		}
		// The whitelist leaves only conversions supported by this helper.
		result, _ = convertScalar(scalar, target, true)
	}
	value := scalarValue(result)
	value.Sources = input.Sources
	if result.k == 'i' || scalar.k == 'i' || scalar.k == 'b' || scalar.k == 'n' {
		value.Safe = everyContext
	} else {
		value.Safe = input.Safe
	}
	return value, true
}

// Equality is established only after every entry agrees. A structural or
// proven scalar mismatch is decisive, even when another primitive is unknown.
func pureStrictRecord(left, right Value, r *scopeResult, depth int) (bool, bool) {
	if !purePlainRecord(left) || !purePlainRecord(right) {
		return false, false
	}
	exhausted := false
	equal, known := strictRecordValues(left, right, r, depth, &exhausted)
	return equal, known && !exhausted
}

func strictRecordValues(left, right Value, r *scopeResult, depth int, exhausted *bool) (bool, bool) {
	if depth == 0 || r.operations >= MaxTransfers {
		*exhausted = true
		return false, false
	}
	if !pureReadOnlyInputs([]Value{left, right}, nil) {
		return false, false
	}
	r.operations++
	a, ak := knownScalar(left)
	b, bk := knownScalar(right)
	if ak && bk {
		equal, _ := constantBinary(syntax.TIsIdentical, a, b)
		return equal.b, true
	}
	ar, br := purePlainRecord(left), purePlainRecord(right)
	if ar && bk || br && ak {
		return false, true
	}
	if !ar || !br || len(left.elements.order) != len(left.elements.items) || len(right.elements.order) != len(right.elements.items) {
		return false, false
	}
	if len(left.elements.order) != len(right.elements.order) {
		return false, true
	}
	uncertain := false
	for i, key := range left.elements.order {
		if key != right.elements.order[i] {
			return false, true
		}
		equal, known := strictRecordValues(left.elements.items[key], right.elements.items[key], r, depth-1, exhausted)
		if *exhausted {
			return false, false
		}
		if known && !equal {
			return false, true
		}
		if !known {
			uncertain = true
		}
	}
	return !uncertain, !uncertain
}
