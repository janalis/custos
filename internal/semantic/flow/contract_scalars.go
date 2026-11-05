package flow

import (
	"math"
	"strconv"
	"strings"

	"custos/internal/php/syntax"
)

func knownScalar(value Value) (constantScalar, bool) {
	if !value.Complete || value.replaced || value.Invalidated != 0 {
		return constantScalar{}, false
	}
	if value.literal != nil {
		return *value.literal, true
	}
	fuel := 256
	return constantValue(value.Expr, MaxDepth, &fuel)
}

func scalarSnapshot(value Value) Value {
	if value.literal == nil && value.Expr != nil {
		if scalar, known := knownScalar(value); known {
			literal := scalar
			value.literal = &literal
		}
	}
	return value
}

func sameScalar(a, b *constantScalar) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil || a.k != b.k {
		return false
	}
	if a.k == 'f' {
		return math.Float64bits(a.f) == math.Float64bits(b.f)
	}
	return *a == *b
}

func scalarExpression(value constantScalar) syntax.Expr {
	switch value.k {
	case 'i':
		return &syntax.Literal{LitKind: syntax.LitInt, Raw: strconv.FormatInt(value.i, 10)}
	case 'f':
		return &syntax.Literal{LitKind: syntax.LitFloat, Raw: strconv.FormatFloat(value.f, 'g', -1, 64)}
	case 's':
		return &syntax.Literal{LitKind: syntax.LitString, Raw: "'" + strings.ReplaceAll(strings.ReplaceAll(value.s, `\`, `\\`), `'`, `\'`) + "'"}
	case 'b':
		name := "false"
		if value.b {
			name = "true"
		}
		return &syntax.ConstFetch{Name: &syntax.Name{Value: name}}
	default:
		return &syntax.ConstFetch{Name: &syntax.Name{Value: "null"}}
	}
}

// Only warning-free primitive conversions are modeled. Numeric strings with
// whitespace, fractions, exponents or overflow remain unknown.
func convertScalar(value constantScalar, target string, weak bool) (constantScalar, bool) {
	if target == "float" && value.k == 'i' && value.i >= -(1<<53) && value.i <= 1<<53 {
		return constantScalar{k: 'f', f: float64(value.i)}, true
	}
	if !weak {
		return constantScalar{}, false
	}
	switch target {
	case "bool":
		if value.k != 'n' {
			return constantScalar{k: 'b', b: value.truth()}, true
		}
	case "int":
		switch value.k {
		case 'b':
			i := int64(0)
			if value.b {
				i = 1
			}
			return constantScalar{k: 'i', i: i}, true
		case 's':
			text := value.s
			if text != "" && (text[0] == '+' || text[0] == '-') {
				text = text[1:]
			}
			if text == "" {
				return constantScalar{}, false
			}
			for _, c := range text {
				if c < '0' || c > '9' {
					return constantScalar{}, false
				}
			}
			i, err := strconv.ParseInt(value.s, 10, 64)
			return constantScalar{k: 'i', i: i}, err == nil
		case 'f':
			if math.Trunc(value.f) == value.f && value.f >= -9223372036854775808.0 && value.f < 9223372036854775808.0 {
				return constantScalar{k: 'i', i: int64(value.f)}, true
			}
		}
	case "float":
		if value.k == 'b' {
			f := float64(0)
			if value.b {
				f = 1
			}
			return constantScalar{k: 'f', f: f}, true
		}
	case "string":
		switch value.k {
		case 'i':
			return constantScalar{k: 's', s: strconv.FormatInt(value.i, 10)}, true
		case 'b':
			text := ""
			if value.b {
				text = "1"
			}
			return constantScalar{k: 's', s: text}, true
		case 'f':
			if value.f == 0 {
				text := "0"
				if math.Signbit(value.f) {
					text = "-0"
				}
				return constantScalar{k: 's', s: text}, true
			}
		}
	}
	return constantScalar{}, false
}

func (e *Env) internalScalarCallback(at syntax.Node) bool {
	name, builtin := e.callName(at)
	return builtin && (name == "array_map" || name == "array_filter" || name == "array_reduce")
}
