package infer

import (
	"math"
	"strconv"
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

func (e *Env) constType(n *syntax.ConstFetch) types.Type {
	return e.resolvedConstType(n)
}

// literalTextType infers the type of a constant initialiser from its source text.
func literalTextType(v string) types.Type {
	v = strings.TrimSpace(v)
	if v == "" {
		return types.Unknown
	}
	switch c := v[0]; {
	case c == '\'' || c == '"':
		return types.String
	case c == '[' || strings.HasPrefix(strings.ToLower(v), "array("):
		return types.Array
	case (c >= '0' && c <= '9') || c == '-' || c == '.':
		if strings.ContainsAny(v, ".eE") && !strings.HasPrefix(strings.ToLower(v), "0x") {
			return types.Float
		}
		return types.Int
	}
	switch strings.ToLower(v) {
	case "true", "false":
		return types.Bool
	case "null":
		return types.Null
	}
	return types.Unknown
}

func (e *Env) unaryType(n *syntax.Unary) types.Type {
	switch n.Op.Kind {
	case syntax.TExclaim:
		return types.Bool
	case syntax.TIntCast:
		return types.Int
	case syntax.TDoubleCast:
		return types.Float
	case syntax.TStringCast:
		return types.String
	case syntax.TArrayCast:
		return castToArray(e.TypeOf(n.Expr))
	case syntax.TBoolCast:
		return types.Bool
	case syntax.TObjectCast:
		return types.Of("object")
	case syntax.TUnsetCast:
		return types.Null
	case syntax.TVoidCast:
		return types.Void
	case syntax.TTilde:
		// ~ on a string flips its bytes and yields a string.
		switch t := e.TypeOf(n.Expr); {
		case t.OnlyOf("string"):
			return types.String
		case mayBeString(t):
			return types.Unknown
		}
		return types.Int
	case syntax.TAt:
		return e.TypeOf(n.Expr)
	default: // TMinus, TPlus: the parser builds no other unary operator
		t := e.TypeOf(n.Expr)
		if t.OnlyOf("int") && n.Op.Kind == syntax.TMinus {
			return negatedInts(t)
		}
		if t.OnlyOf("int") || t.OnlyOf("float") {
			return t
		}
		return types.Unknown
	}
}

// literalType types a literal with its value facts: an integer holds its
// value, a string its value and byte length (a float, or an integer or
// string that does not decode, gets the plain type).
func (e *Env) literalType(n *syntax.Literal) types.Type {
	switch n.LitKind {
	case syntax.LitInt:
		if v, err := strconv.ParseInt(strings.ReplaceAll(n.Raw, "_", ""), 0, 64); err == nil {
			return types.IntValue(v)
		}
		return types.Int
	case syntax.LitFloat:
		return types.Float
	}
	if v, ok := e.literalValue(n); ok {
		return types.StringLiteral(v)
	}
	return types.String
}

// literalValue is string literal n's value: a single-quoted string without
// escapes is read directly, others decoded.
func (e *Env) literalValue(n *syntax.Literal) (string, bool) {
	raw := n.Raw
	if raw != "" && (raw[0] == 'b' || raw[0] == 'B') {
		raw = raw[1:]
	}
	if len(raw) >= 2 && raw[0] == '\'' && raw[len(raw)-1] == '\'' && strings.IndexByte(raw, '\\') < 0 {
		return raw[1 : len(raw)-1], true
	}
	return syntax.LiteralStringValueForVersion(n, e.PHP)
}

// negatedInts is the int type -x holds for x of int type t: its values
// negated when known (a plain int when one is the smallest int, whose
// negation is a float).
func negatedInts(t types.Type) types.Type {
	rs, ok := t.IntRanges()
	if !ok {
		return t
	}
	out := make([]types.IntRange, len(rs))
	for i, r := range rs {
		if r.Lo == math.MinInt64 {
			return types.Int
		}
		out[i] = types.IntRange{Lo: -r.Hi, Hi: -r.Lo}
	}
	return types.Ints(out...)
}

// numeric types `a + b`, `a - b` and `a * b` on numbers. One numeric
// operand is enough: whatever the other holds, the result is a number
// (null, booleans and numeric strings convert; arrays and objects throw on
// PHP 8 and convert with a notice before), a float with a float operand.
func (e *Env) numeric(a, b syntax.Expr) types.Type {
	ta, tb := e.TypeOf(a), e.TypeOf(b)
	switch {
	case ta.OnlyOf("int") && tb.OnlyOf("int"):
		return types.Int
	case ta.OnlyOf("float") || tb.OnlyOf("float"):
		return types.Float
	case ta.OnlyOf("int", "float") || tb.OnlyOf("int", "float"):
		return types.Of("int", "float")
	}
	return types.Unknown
}

func (e *Env) binaryType(n *syntax.Binary) types.Type {
	if t, ok := e.overloaded(n.Op.Kind, n.Left, n.Right); ok {
		return t
	}
	switch n.Op.Kind {
	case syntax.TPipe:
		return e.pipeType(n.Left, n.Right)
	case syntax.TDot:
		return types.String
	case syntax.TPlus:
		if a, b := e.TypeOf(n.Left), e.TypeOf(n.Right); a.IsArrayLike() && b.IsArrayLike() {
			return arrayUnion(a, b)
		}
		return e.numeric(n.Left, n.Right)
	case syntax.TPow:
		return e.powerType(n.Left, n.Right)
	case syntax.TMinus, syntax.TMul:
		return e.numeric(n.Left, n.Right)
	case syntax.TDiv:
		if t := e.numeric(n.Left, n.Right); !t.IsUnknown() {
			return types.Of("int", "float")
		}
		return types.Unknown
	case syntax.TAmpersand, syntax.TBar, syntax.TCaret:
		return e.bitwise(n.Left, n.Right)
	case syntax.TMod, syntax.TSl, syntax.TSr, syntax.TSpaceship:
		return types.Int
	case syntax.TCoalesce:
		return types.Union(e.TypeOf(n.Left).Without("null"), e.TypeOf(n.Right))
	default: // All remaining parser binary operators are comparisons or boolean operators.
		return types.Bool
	}
}

func (e *Env) compoundType(n *syntax.Assign) types.Type {
	if t, ok := e.overloaded(n.Op.Kind, n.Var, n.Value); ok {
		return t
	}
	switch n.Op.Kind {
	case syntax.TConcatEqual:
		return types.String
	case syntax.TPlusEqual:
		if a, b := e.TypeOf(n.Var), e.TypeOf(n.Value); a.IsArrayLike() && b.IsArrayLike() {
			return arrayUnion(a, b)
		}
		return e.numeric(n.Var, n.Value)
	case syntax.TPowEqual:
		return e.powerType(n.Var, n.Value)
	case syntax.TMinusEqual, syntax.TMulEqual:
		return e.numeric(n.Var, n.Value)
	case syntax.TDivEqual: // as `/`
		if t := e.numeric(n.Var, n.Value); !t.IsUnknown() {
			return types.Of("int", "float")
		}
		return types.Unknown
	case syntax.TCoalesceEqual:
		lt := e.TypeOf(n.Var)
		if d, ok := syntax.UnwrapParens(n.Var).(*syntax.ArrayDimFetch); ok && lt.IsUnknown() {
			// `$by[$k] ??= []` on a literal array no write has filled yet:
			// the key is absent, the result is the new value.
			if v := asVariable(d.Var); v != nil && d.Dim != nil {
				if ct := e.baseType(v); ct.IsSealedShape() && len(ct.ShapeKeys()) == 0 {
					if ws, _ := e.reachingWrites(v); len(ws) == 0 {
						return e.TypeOf(n.Value)
					}
				}
			}
		}
		return types.Union(lt.Without("null"), e.TypeOf(n.Value))
	case syntax.TAndEqual, syntax.TOrEqual, syntax.TXorEqual:
		return e.bitwise(n.Var, n.Value)
	default: // %=, <<=, >>= (syntax.TokenKind.IsAssignOp lists every compound operator)
		return types.Int
	}
}

// overloaded types an operator whose operand is an object overloading it:
// \GMP answers arithmetic, shift and bitwise operators with a \GMP, and
// \BcMath\Number answers arithmetic operators with a \BcMath\Number (both
// classes are final), whatever the other operand.
func (e *Env) overloaded(op syntax.TokenKind, a, b syntax.Expr) (types.Type, bool) {
	var classes []string
	switch op {
	case syntax.TPlus, syntax.TMinus, syntax.TMul, syntax.TDiv, syntax.TMod, syntax.TPow,
		syntax.TPlusEqual, syntax.TMinusEqual, syntax.TMulEqual, syntax.TDivEqual, syntax.TModEqual, syntax.TPowEqual:
		classes = overloadingArithmetic
	case syntax.TSl, syntax.TSr, syntax.TAmpersand, syntax.TBar, syntax.TCaret,
		syntax.TSlEqual, syntax.TSrEqual, syntax.TAndEqual, syntax.TOrEqual, syntax.TXorEqual:
		classes = overloadingBitwise
	default:
		return types.Type{}, false
	}
	for _, x := range [2]syntax.Expr{a, b} {
		t := e.TypeOf(x)
		for _, c := range classes {
			if t.OnlyOf(c) {
				return t, true
			}
		}
	}
	return types.Type{}, false
}

var (
	overloadingArithmetic = []string{`\GMP`, `\BcMath\Number`}
	overloadingBitwise    = []string{`\GMP`}
)

// bitwise types `&`, `|`, `^`: a string when both operands are strings
// (PHP operates on the bytes), else an int; unknown when both may be strings.
func (e *Env) bitwise(a, b syntax.Expr) types.Type {
	ta, tb := e.TypeOf(a), e.TypeOf(b)
	switch {
	case ta.OnlyOf("string") && tb.OnlyOf("string"):
		return types.String
	case mayBeString(ta) && mayBeString(tb):
		return types.Unknown
	}
	return types.Int
}

// mayBeString reports a type that is unknown or may hold a string.
func mayBeString(t types.Type) bool { return t.IsUnknown() || t.HasAny("string", "mixed") }

// powerType keeps integers only when literal operands prove no overflow.
func (e *Env) powerType(a, b syntax.Expr) types.Type {
	if e.numeric(a, b).IsUnknown() {
		return types.Unknown
	}
	if powerFitsInt(a, b) {
		return types.Int
	}
	return types.Of("int", "float")
}

func powerFitsInt(a, b syntax.Expr) bool {
	x, xok := integerLiteral(a)
	y, yok := integerLiteral(b)
	if !xok || !yok || y < 0 {
		return false
	}
	if y == 0 || x == 0 || x == 1 || x == -1 {
		return true
	}
	if y > 63 {
		return false
	}
	result := int64(1)
	for range y {
		next := result * x
		if next/x != result || next == -1<<63 && x == -1 {
			return false
		}
		result = next
	}
	return true
}

func integerLiteral(x syntax.Expr) (int64, bool) {
	x = syntax.UnwrapParens(x)
	if u, ok := x.(*syntax.Unary); ok && u.Op.Kind == syntax.TMinus {
		v, valid := integerLiteral(u.Expr)
		return -v, valid && v != -1<<63
	}
	lit, ok := x.(*syntax.Literal)
	if !ok || lit.LitKind != syntax.LitInt {
		return 0, false
	}
	v, err := strconv.ParseInt(strings.ReplaceAll(lit.Raw, "_", ""), 0, 64)
	return v, err == nil
}

// castToArray types `(array) $x` for $x of type t: arrays stay as they
// are, null becomes an empty array and a scalar a list of itself; an
// object (its properties) or anything unknown gives an array.
func castToArray(t types.Type) types.Type {
	if t.IsUnknown() {
		return types.Array
	}
	var others, scalars []string
	parts := []types.Type{}
	for _, a := range t.Atoms() {
		switch a {
		case "null":
			parts = append(parts, types.FromDoc("array{}", nil))
		case "int", "float", "string", "bool", "false", "true":
			scalars = append(scalars, a)
		default:
			if a != "array" && !strings.HasSuffix(a, "[]") {
				return types.Array
			}
			continue
		}
		others = append(others, a)
	}
	if arrays := t.Without(others...); len(arrays.Atoms()) > 0 {
		parts = append(parts, arrays)
	}
	if len(scalars) > 0 {
		parts = append(parts, listOf(types.Of(scalars...), true))
	}
	return types.Union(parts...)
}
