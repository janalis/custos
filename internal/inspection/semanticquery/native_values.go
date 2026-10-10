package semanticquery

import (
	"strconv"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// NativeArrayEntries evaluates literal array keys with PHP's overwrite and
// next-index behavior. Unknown keys do not yield a fabricated element count.
func NativeArrayEntries(ctx *analysis.Context, e syntax.Expr) (map[string]syntax.Expr, bool) {
	a := NativeArray(ctx, e)
	if a == nil {
		return nil, false
	}
	entries := make(map[string]syntax.Expr, len(a.Items))
	next := int64(0)
	for _, item := range a.Items {
		key := "i:" + strconv.FormatInt(next, 10)
		if item.Key != nil {
			var ok bool
			key, ok = NativeArrayKey(ctx, item.Key)
			if !ok {
				return nil, false
			}
		}
		if key[:2] == "i:" {
			n, _ := strconv.ParseInt(key[2:], 10, 64)
			if n >= next || (len(entries) == 0 && ctx.PHP >= phpversion.PHP83) {
				if n == 1<<63-1 {
					return nil, false
				}
				next = n + 1
			}
		}
		entries[key] = item.Value
	}
	return entries, true
}

// NativeArrayKey canonicalizes known integer/string keys using PHP's decimal
// string conversion. Other key types remain unknown to avoid coercion guesses.
func NativeArrayKey(ctx *analysis.Context, e syntax.Expr) (string, bool) {
	if n, ok := NativeInt(ctx, e); ok {
		return "i:" + strconv.FormatInt(n, 10), true
	}
	if s, ok := NativeString(ctx, e); ok {
		if n, err := strconv.ParseInt(s, 10, 64); err == nil && strconv.FormatInt(n, 10) == s {
			return "i:" + s, true
		}
		return "s:" + s, true
	}
	return "", false
}

// CallArgument finds a positional or named argument. Spreads and duplicate
// bindings are deliberately unknown: inspections must not guess their values.
func CallArgument(list *syntax.ArgList, position int, name string) syntax.Expr {
	if list == nil {
		return nil
	}
	var result syntax.Expr
	pos := 0
	for _, node := range list.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok || arg.Unpack {
			return nil
		}
		matches := arg.Name != nil && arg.Name.Value == name
		if arg.Name == nil {
			matches = pos == position
			pos++
		}
		if matches {
			if result != nil {
				return nil
			}
			result = arg.Value
		}
	}
	return result
}

// NativeValue returns a single discovered expression, retaining the distinction
// between unknown values and literals. The discovery procedure is bounded.
func NativeValue(ctx *analysis.Context, e syntax.Expr) syntax.Expr {
	if e == nil {
		return nil
	}
	if _, ok := syntax.UnwrapParens(e).(*syntax.Variable); ok {
		v, known := ctx.Flow().Resolve(e)
		if !known {
			return nil
		}
		return syntax.UnwrapParens(v)
	}
	values, known := DiscoverValuesKnown(ctx.Types(), e)
	if !known || len(values) != 1 {
		return nil
	}
	return syntax.UnwrapParens(values[0])
}

// NativeArray resolves a literal array with no spread or referenced items.
// Such arrays have a statically known number of entries and stable values.
func NativeArray(ctx *analysis.Context, e syntax.Expr) *syntax.Array {
	a, ok := NativeValue(ctx, e).(*syntax.Array)
	if !ok {
		return nil
	}
	for _, item := range a.Items {
		if item == nil || item.Unpack || item.ByRef {
			return nil
		}
	}
	return a
}

// NativeString returns a decoded single string literal, not interpolated text.
func NativeString(ctx *analysis.Context, e syntax.Expr) (string, bool) {
	return astquery.QuotedStringValue(NativeValue(ctx, e))
}

// NativeInt returns a known integer literal, including its unary sign.
func NativeInt(ctx *analysis.Context, e syntax.Expr) (int64, bool) {
	e = NativeValue(ctx, e)
	if constant, ok := e.(*syntax.ConstFetch); ok {
		if resolved := ResolveConstant(ctx.Types(), constant); resolved != nil {
			return astquery.ParseIntLiteral(resolved.Value)
		}
	}
	negative := false
	if unary, ok := e.(*syntax.Unary); ok {
		if unary.Op.Kind != syntax.TMinus && unary.Op.Kind != syntax.TPlus {
			return 0, false
		}
		negative = unary.Op.Kind == syntax.TMinus
		e = syntax.UnwrapParens(unary.Expr)
	}
	lit, ok := e.(*syntax.Literal)
	if !ok || lit.LitKind != syntax.LitInt {
		return 0, false
	}
	n, ok := astquery.ParseIntLiteral(lit.Raw)
	if negative {
		n = -n
	}
	return n, ok
}

// NativeFlag proves the presence or absence of a bit in a known flag value.
// Dynamic values remain unknown even when a nearby literal looks reassuring.
func NativeFlag(ctx *analysis.Context, e syntax.Expr, mask int64) (bool, bool) {
	if e == nil {
		return false, true
	}
	if n, ok := NativeInt(ctx, e); ok {
		return n&mask != 0, true
	}
	if binary, ok := NativeValue(ctx, e).(*syntax.Binary); ok && binary.Op.Kind == syntax.TBar {
		a, ak := NativeFlag(ctx, binary.Left, mask)
		b, bk := NativeFlag(ctx, binary.Right, mask)
		return a || b, ak && bk
	}
	return false, false
}

// NativeTruth returns PHP truthiness only for known scalar literals.
func NativeTruth(ctx *analysis.Context, e syntax.Expr) (bool, bool) {
	e = NativeValue(ctx, e)
	if b, ok := astquery.BoolConst(e); ok {
		return b, true
	}
	if n, ok := NativeInt(ctx, e); ok {
		return n != 0, true
	}
	if s, ok := NativeString(ctx, e); ok {
		return s != "" && s != "0", true
	}
	if c, ok := e.(*syntax.ConstFetch); ok && c.Name != nil && c.Name.Value == "null" {
		return false, true
	}
	if lit, ok := e.(*syntax.Literal); ok && lit.LitKind == syntax.LitFloat {
		v, err := strconv.ParseFloat(lit.Raw, 64)
		return v != 0, err == nil
	}
	return false, false
}
