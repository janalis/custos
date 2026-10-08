package probablebugs

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// pregQuoteUsage reports preg_quote() calls without the delimiter argument.
type pregQuoteUsage struct{}

func init() { register(pregQuoteUsage{}) }

// Semantic marks the rule as needing the project index (a user function
// named preg_quote may be declared in another file).
func (pregQuoteUsage) Semantic() {}

func (pregQuoteUsage) ID() string { return "PregQuoteUsage" }

func (pregQuoteUsage) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func (pregQuoteUsage) Check(ctx *analysis.Context, n syntax.Node) {
	call, ok := util.IsFuncNamedFold(n, "preg_quote")        // D1
	if !ok || call.Args == nil || len(call.Args.Args) != 1 { // D2
		return
	}
	if _, ok := call.Args.Args[0].(*syntax.Arg); !ok {
		return // first-class callable syntax preg_quote(...)
	}
	if !util.ResolvesToGlobalFunction(ctx.Names(), ctx.Index(), ctx.PHP, call, "preg_quote") { // D1
		return
	}
	if pregQuoteDelimiterEscaped(ctx, call) || pregQuoteDelimiterFree(ctx, call) { // E3
		return
	}
	ctx.Report(util.NamePartSpan(call.Name.(*syntax.Name)), "Pass the pattern delimiter to preg_quote() as its second argument.")
}

// pregQuoteDelimiterEscaped reports whether the pattern the call is
// concatenated into (or the sprintf() format it is an argument of) visibly
// starts with a delimiter that preg_quote() escapes anyway: the first
// non-blank character of the leftmost string literal of the concatenation.
func pregQuoteDelimiterEscaped(ctx *analysis.Context, call *syntax.FuncCall) bool {
	var top syntax.Node = call
	for {
		parent, child := util.ParentSkipParens(top)
		if b, ok := parent.(*syntax.Binary); ok && b.Op.Kind == syntax.TDot {
			top = b
			continue
		}
		// str_replace('\\*', '.*', preg_quote($glob)) is still the quoted
		// text, rewritten (custos refinement).
		if outer := pregQuoteRewriter(ctx, child); outer != nil {
			top = outer
			continue
		}
		top = child
		break
	}
	var first syntax.Expr
	if b, ok := top.(*syntax.Binary); ok {
		first = b.Left
		for {
			inner, ok := syntax.UnwrapParens(first).(*syntax.Binary)
			if !ok || inner.Op.Kind != syntax.TDot {
				break
			}
			first = inner.Left
		}
	} else if arg, ok := top.Parent().(*syntax.Arg); ok {
		// sprintf('|^%s$|', preg_quote($x))
		args := arg.Parent().(*syntax.ArgList)
		outer, ok := args.Parent().(*syntax.FuncCall)
		if !ok || ctx.GlobalFunctionName(outer) != "sprintf" || len(args.Args) == 0 || args.Args[0] == syntax.Node(arg) {
			return false
		}
		if fa, ok := args.Args[0].(*syntax.Arg); ok {
			first = fa.Value
		}
	}
	if first == nil {
		return false
	}
	v, ok := util.QuotedStringValue(syntax.UnwrapParens(first))
	v = strings.TrimLeft(v, " \t\n\r\v\f")
	if !ok || v == "" {
		return false
	}
	return strings.IndexByte(pregQuoteEscapes(ctx), v[0]) >= 0
}

// pregQuoteEscapes lists the characters preg_quote() escapes by itself.
func pregQuoteEscapes(ctx *analysis.Context) string {
	if ctx.PHP >= phpver.PHP73 {
		return `.\+*?[^]$(){}=!<>|:-#`
	}
	return `.\+*?[^]$(){}=!<>|:-`
}

// pregQuoteDelimiterFree reports whether the quoted text is a plain string
// literal whose result cannot contain an unescaped delimiter: every
// character is alphanumeric, whitespace or a backslash (none of which can
// be a delimiter) or one preg_quote() escapes anyway (custos refinement:
// preg_quote('::'), preg_quote('\\')).
func pregQuoteDelimiterFree(ctx *analysis.Context, call *syntax.FuncCall) bool {
	arg := syntax.UnwrapParens(call.Args.Args[0].(*syntax.Arg).Value)
	vals := []syntax.Expr{arg}
	switch arg.(type) {
	case *syntax.ClassConstFetch, *syntax.ConstFetch:
		// custos: a constant holding such a literal (`Packer::PREFIX`).
		found, known := util.DiscoverValuesKnown(ctx.Types(), arg)
		if !known || len(found) == 0 {
			return false
		}
		vals = found
	}
	escaped := pregQuoteEscapes(ctx)
	for _, e := range vals {
		v, ok := util.QuotedStringValue(e)
		if !ok {
			return false
		}
		for i := 0; i < len(v); i++ {
			c := v[i]
			switch {
			case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9',
				c == ' ', c == '\t', c == '\n', c == '\r', c == '\v', c == '\f':
			case strings.IndexByte(escaped, c) >= 0:
			default:
				return false
			}
		}
	}
	return true
}

// pregQuoteRewriter returns the str_replace() call whose subject (third
// argument), or the strtr() call whose string (first argument), is e.
func pregQuoteRewriter(ctx *analysis.Context, e syntax.Node) *syntax.FuncCall {
	arg, ok := e.Parent().(*syntax.Arg)
	if !ok || arg.Unpack || arg.Name != nil {
		return nil
	}
	args := arg.Parent().(*syntax.ArgList)
	outer, ok := args.Parent().(*syntax.FuncCall)
	if !ok {
		return nil
	}
	want := -1
	switch ctx.GlobalFunctionName(outer) {
	case "str_replace":
		want = 2
	case "strtr":
		want = 0
	}
	if want < 0 || len(args.Args) <= want || args.Args[want] != syntax.Node(arg) {
		return nil
	}
	return outer
}
