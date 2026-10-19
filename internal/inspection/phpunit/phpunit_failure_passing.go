package phpunit

import (
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// FailurePassing returns the arguments of a PHPUnit assertion that pass when
// they hold a failure value false or null as if it were text: a value
// loosely compared with a literal that equals false, a value asserted empty
// and, with textual (a failure false that a string parameter coerces to the
// empty text when the file does not declare strict types), a needle, prefix
// or suffix, which the empty text matches in any text.
func FailurePassing(ctx *analysis.Context, call syntax.Node, args *syntax.ArgList, textual bool) []syntax.Expr {
	// PuAssertion resolved the function, or read a literal method name.
	var name string
	if f, ok := call.(*syntax.FuncCall); ok {
		fqn := ctx.Types().ResolveFunction(f).FQN
		name = fqn[strings.LastIndexByte(fqn, '\\')+1:]
	} else {
		c, _ := AsPuCall(call)
		name = c.Ident.Value
	}
	needle := textual && !syntax.StrictTypes(ctx.File)
	switch strings.ToLower(name) {
	case "assertstringcontainsstring", "assertstringcontainsstringignoringcase":
		if needle {
			return passed(args, 0, "needle")
		}
	case "assertstringstartswith":
		if needle {
			return passed(args, 0, "prefix")
		}
	case "assertstringendswith":
		if needle {
			return passed(args, 0, "suffix")
		}
	case "assertequals", "assertequalscanonicalizing", "assertequalsignoringcase":
		expected, actual := argument(args, 0, "expected"), argument(args, 1, "actual")
		if falsy(expected, textual) && actual != nil {
			return []syntax.Expr{actual}
		}
		if falsy(actual, textual) && expected != nil {
			return []syntax.Expr{expected}
		}
	case "assertempty":
		return passed(args, 0, "actual")
	}
	return nil
}

func passed(args *syntax.ArgList, position int, name string) []syntax.Expr {
	if value := argument(args, position, name); value != nil {
		return []syntax.Expr{value}
	}
	return nil
}

// argument returns the value bound to a parameter, positionally or by name;
// nothing when an unpacked list or a placeholder hides the binding.
func argument(args *syntax.ArgList, position int, name string) syntax.Expr {
	var bound syntax.Expr
	at := 0
	for _, node := range args.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok || arg.Unpack {
			return nil
		}
		switch {
		case arg.Name != nil:
			if arg.Name.Value == name {
				bound = arg.Value
			}
		case at == position:
			bound = arg.Value
			at++
		default:
			at++
		}
	}
	return bound
}

// falsy reports a literal that equals the failure value loosely: false,
// null, the empty text or zero, and "0" for a failure false (null is
// compared with text as the empty text).
func falsy(e syntax.Expr, failsFalse bool) bool {
	switch x := syntax.UnwrapParens(e).(type) {
	case *syntax.ConstFetch:
		return x.Name != nil && (strings.EqualFold(strings.TrimPrefix(x.Name.Value, `\`), "false") || strings.EqualFold(strings.TrimPrefix(x.Name.Value, `\`), "null"))
	case *syntax.Literal:
		switch x.LitKind {
		case syntax.LitString:
			text, ok := syntax.LiteralStringValue(x)
			return ok && (text == "" || failsFalse && text == "0")
		case syntax.LitInt, syntax.LitFloat:
			v, err := strconv.ParseFloat(strings.ReplaceAll(x.Raw, "_", ""), 64)
			return err == nil && v == 0
		}
	}
	return false
}

// webmozartAssert names the assertion class of webmozart/assert, matched by
// the name written in the call: its source need not be analysed.
const webmozartAssert = `webmozart\assert\assert`

// LibraryFailurePassing returns, for a static call of an assertion of
// webmozart/assert, the arguments that pass when they hold a failure value
// as if it were a value: a value loosely compared by `eq` with a literal that
// equals the failure, a value asserted `isEmpty`, and, for a failure null
// (textual false), the value of any `nullOr…` assertion, which accepts null.
// ok is false for any other call.
func LibraryFailurePassing(ctx *analysis.Context, call syntax.Node, args *syntax.ArgList, textual bool) (inputs []syntax.Expr, ok bool) {
	static, isStatic := call.(*syntax.StaticCall)
	if !isStatic {
		return nil, false
	}
	class, named := static.Class.(*syntax.Name)
	method, literal := static.Name.(*syntax.Identifier)
	if !named || !literal || !strings.EqualFold(strings.TrimPrefix(PuResolveClassName(ctx, class), `\`), webmozartAssert) {
		return nil, false
	}
	name := strings.ToLower(method.Value)
	if after, nullable := strings.CutPrefix(name, "nullor"); nullable {
		if !textual {
			return passed(args, 0, "value"), true
		}
		name = after
	}
	switch name {
	case "eq":
		if value := argument(args, 0, "value"); value != nil && falsy(argument(args, 1, "expect"), textual) {
			return []syntax.Expr{value}, true
		}
	case "isempty":
		return passed(args, 0, "value"), true
	}
	return nil, true
}
