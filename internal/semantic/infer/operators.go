package infer

import (
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
		return types.Array
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
		if t.OnlyOf("int") || t.OnlyOf("float") {
			return t
		}
		return types.Unknown
	}
}

func (e *Env) numeric(a, b syntax.Expr) types.Type {
	ta, tb := e.TypeOf(a), e.TypeOf(b)
	switch {
	case ta.OnlyOf("int") && tb.OnlyOf("int"):
		return types.Int
	case ta.OnlyOf("int", "float") && tb.OnlyOf("int", "float"):
		return types.Float
	}
	return types.Unknown
}

func (e *Env) binaryType(n *syntax.Binary) types.Type {
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
	case syntax.TMinus, syntax.TMul, syntax.TPow:
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
	switch n.Op.Kind {
	case syntax.TConcatEqual:
		return types.String
	case syntax.TPlusEqual:
		if a, b := e.TypeOf(n.Var), e.TypeOf(n.Value); a.IsArrayLike() && b.IsArrayLike() {
			return arrayUnion(a, b)
		}
		return e.numeric(n.Var, n.Value)
	case syntax.TMinusEqual, syntax.TMulEqual, syntax.TPowEqual:
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
