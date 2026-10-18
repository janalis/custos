package semanticquery

import (
	"strconv"
	"strings"

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

// ArgumentState distinguishes an omitted parameter from an uncertain binding.
type ArgumentState uint8

const (
	ArgumentMissing ArgumentState = iota
	ArgumentBound
	ArgumentUnknown
)

// CallArgumentBinding finds a positional or named argument. Unpacking,
// placeholders and duplicate bindings cannot establish absence or a value.
// Names from the index may include the parameter's leading dollar sign.
func CallArgumentBinding(list *syntax.ArgList, position int, name string) (syntax.Expr, ArgumentState) {
	if list == nil {
		return nil, ArgumentMissing
	}
	name = strings.TrimPrefix(name, "$")
	var result syntax.Expr
	pos := 0
	for _, node := range list.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok || arg.Unpack {
			return nil, ArgumentUnknown
		}
		matches := arg.Name != nil && arg.Name.Value == name
		if arg.Name == nil {
			matches = pos == position
			pos++
		}
		if matches {
			if result != nil {
				return nil, ArgumentUnknown
			}
			result = arg.Value
		}
	}
	if result == nil {
		return nil, ArgumentMissing
	}
	return result, ArgumentBound
}

// CallArgumentOmitted reports whether the parameter is provably not passed,
// so that PHP uses its default value.
func CallArgumentOmitted(list *syntax.ArgList, position int, name string) bool {
	_, state := CallArgumentBinding(list, position, name)
	return state == ArgumentMissing
}

// CallArgument returns a proven argument value. Callers requiring proof of
// omission must use CallArgumentBinding instead of comparing this with nil.
func CallArgument(list *syntax.ArgList, position int, name string) syntax.Expr {
	value, _ := CallArgumentBinding(list, position, name)
	return value
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
		return NativeUnsilenced(v)
	}
	values, known := DiscoverValuesKnown(ctx.Types(), e)
	if !known || len(values) != 1 {
		return nil
	}
	return NativeUnsilenced(values[0])
}

// NativeUnsilenced looks through parentheses and the `@` operator: silencing
// a call hides its diagnostics, its result is the result of the call.
func NativeUnsilenced(e syntax.Expr) syntax.Expr {
	for {
		e = syntax.UnwrapParens(e)
		silenced, ok := e.(*syntax.Unary)
		if !ok || silenced.Op.Kind != syntax.TAt {
			return e
		}
		e = silenced.Expr
	}
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

// nativeStepBudget bounds the chain of steps followed back to a literal.
const nativeStepBudget = 16

// NativeInt returns a known integer literal, including its unary sign. A
// variable that held a known integer and was then stepped with `++`/`--` or
// `+=`/`-=` a known integer, on the only path to the read, has the stepped
// value.
func NativeInt(ctx *analysis.Context, e syntax.Expr) (int64, bool) {
	return nativeInt(ctx, e, nativeStepBudget)
}

func nativeInt(ctx *analysis.Context, e syntax.Expr, budget int) (int64, bool) {
	_, variable := syntax.UnwrapParens(e).(*syntax.Variable)
	e = NativeValue(ctx, e)
	if variable && budget > 0 {
		// The flow value of a stepped variable is the step itself; the
		// operand it reads still holds the value before the step.
		var before syntax.Expr
		delta, stepped := int64(0), false
		switch step := e.(type) {
		case *syntax.IncDec:
			before, delta, stepped = step.Var, 1, true
			if step.Op.Kind == syntax.TDec {
				delta = -1
			}
		case *syntax.Assign:
			if step.Op.Kind == syntax.TPlusEqual || step.Op.Kind == syntax.TMinusEqual {
				before = step.Var
				// A known operand is never the minimum integer, which PHP
				// reads as a float, so its negation cannot overflow.
				delta, stepped = nativeInt(ctx, step.Value, budget-1)
				if step.Op.Kind == syntax.TMinusEqual {
					delta = -delta
				}
			}
		}
		if stepped {
			base, known := nativeInt(ctx, before, budget-1)
			sum := base + delta
			// An overflowing sum is a float in PHP.
			if !known || (delta > 0) != (sum > base) && delta != 0 {
				return 0, false
			}
			return sum, true
		}
	}
	if constant, ok := e.(*syntax.ConstFetch); ok {
		if resolved := ResolveConstant(ctx.Types(), constant); resolved != nil {
			// A declared value carries its sign, which the literal parser rejects.
			n, ok := astquery.ParseIntLiteral(strings.TrimPrefix(resolved.Value, "-"))
			if strings.HasPrefix(resolved.Value, "-") {
				n = -n
			}
			return n, ok
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
