package substrshorthandusage

import (
	"slices"
	"strconv"
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// subStrShortHandUsage reports substr()/mb_substr() lengths computed as
// `strlen($s) - x` that are a negative constant or unnecessary.
type subStrShortHandUsage struct{}

func (subStrShortHandUsage) ID() string               { return "SubStrShortHandUsage" }
func (subStrShortHandUsage) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (subStrShortHandUsage) Check(ctx *analysis.Context, n syntax.Node) {
	call, name := semanticquery.GlobalCall(ctx, n, "substr", "mb_substr") // D1
	if call == nil {
		return
	}
	args, ok := astquery.ArgValues(call.Args)
	if !ok || len(args) < 3 || len(args) > 4 { // D2
		return
	}
	if len(args) == 4 && name == "substr" { // E1b: substr has no 4th parameter
		return
	}
	// E1c: a one-argument mb_strlen counts in the internal encoding (UTF-8
	// by default), not in an explicitly named other one.
	if len(args) == 4 && otherEncoding(ctx, args[3]) {
		return
	}
	subject, start, length := args[0], args[1], args[2]
	sub, ok := length.(*syntax.Binary) // D3
	if !ok || sub.Op.Kind != syntax.TMinus {
		return
	}
	// D4: the length function must count in the same unit as the substr
	// variant (bytes for substr, characters for mb_substr).
	lenName := "strlen"
	if name == "mb_substr" {
		lenName = "mb_strlen"
	}
	strlen, _ := semanticquery.GlobalCall(ctx, sub.Left, lenName)
	if strlen == nil {
		return
	}
	largs, ok := astquery.ArgValues(strlen.Args)
	if !ok || len(largs) != 1 || !astquery.EquivalentFoldNames(ctx.File, largs[0], subject) {
		return
	}
	if lit, ok := start.(*syntax.Literal); ok && lit.LitKind == syntax.LitFloat {
		// E5b: `strlen($s) - 1.5` is truncated after the subtraction, the
		// start before it, so the length does not reach the end.
		return
	}
	span := length.Span()
	// Both rewrites evaluate the subject (and, for Drop, the start) once
	// instead of twice: only offer them when that cannot be observed.
	safe := repeatable(ctx, subject) && repeatable(ctx, start)
	if !astquery.EquivalentFoldNames(ctx.File, sub.Right, start) { // D5
		s, ok1 := decimalInt(ctx, start)
		r, ok2 := decimalInt(ctx, sub.Right)
		if !ok1 || !ok2 || s < 0 { // D6 (negative start: spec Divergences)
			return
		}
		if d := s - r; d < 0 {
			if d < -2 {
				// For a string shorter than R the original length is
				// negative and still keeps characters; `d` keeps none.
				// Some length L in ((R+start)/2, R) exists once d <= -3.
				return
			}
			repl := strconv.FormatInt(d, 10)
			msg := "Pass '" + repl + "' as the length instead."
			if !safe {
				ctx.Report(span, msg)
				return
			}
			ctx.Report(span, msg, astquery.ReplaceFix(span, repl))
			return
		}
	}
	// Drop (F2/F3).
	msg := "The length '" + ctx.Text(length) + "' is unnecessary; remove it."
	if !safe {
		ctx.Report(span, msg)
		return
	}
	first, last := args[0].Span(), args[len(args)-1].Span()
	repl := ctx.Text(subject) + ", " + ctx.Text(start)
	if len(args) == 4 {
		repl += ", null, " + ctx.Text(args[3])
	}
	list := syntax.Span{Start: first.Start, End: last.End}
	ctx.Report(span, msg, diagnostic.Fix{
		Title: "Remove the length",
		Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: list, NewText: repl}} },
	})
}

// otherEncodings are lower-case prefixes of the mbstring encoding names (and
// aliases) that do not count characters the way UTF-8 does.
var otherEncodings = []string{
	"8bit", "binary", "ascii", "us-ascii", "utf-16", "utf16", "utf-32", "utf32", "utf-7", "utf7",
	"ucs-", "iso-", "iso8859", "latin", "windows-", "cp", "sjis", "shift_jis", "euc", "jis", "big5",
	"big-5", "gb", "koi8", "hz", "base64", "html", "quoted-printable", "uuencode", "armscii", "byte",
}

// otherEncoding reports whether e is a string literal naming an encoding
// other than UTF-8. Anything else (a variable, a constant, a name mbstring
// does not know) is taken to be the internal encoding.
func otherEncoding(ctx *analysis.Context, e syntax.Expr) bool {
	v, _ := astquery.StringLiteralValue(ctx.Text(e))
	v = strings.ToLower(v)
	return slices.ContainsFunc(otherEncodings, func(p string) bool { return strings.HasPrefix(v, p) })
}

// decimalInt parses an integer literal, optionally negated, whose source text
// is a plain 32-bit decimal number (`- 2`, `0x2`, `1_0`, `1.0` fail).
func decimalInt(ctx *analysis.Context, e syntax.Expr) (int64, bool) {
	lit := e
	if u, ok := e.(*syntax.Unary); ok && u.Op.Kind == syntax.TMinus {
		lit = u.Expr
	}
	if l, ok := lit.(*syntax.Literal); !ok || l.LitKind != syntax.LitInt {
		return 0, false
	}
	v, err := strconv.ParseInt(ctx.Text(e), 10, 32)
	return v, err == nil
}

// pureFunctions are builtins whose result depends on their arguments only and
// that modify nothing, so evaluating such a call once or twice is the same.
var pureFunctions = []string{
	"strlen", "mb_strlen", "count", "sizeof", "strtolower", "strtoupper", "mb_strtolower", "mb_strtoupper",
	"trim", "ltrim", "rtrim", "ucfirst", "lcfirst", "strrev", "str_repeat", "substr", "mb_substr",
	"strpos", "stripos", "strrpos", "strripos", "mb_strpos", "mb_stripos", "mb_strrpos", "mb_strripos",
	"abs", "min", "max", "intval", "strval", "basename", "dirname", "implode", "join",
}

// repeatable reports whether evaluating e a second time is unobservable: it
// contains no write, object creation, control transfer or call other than a
// pure builtin. Closure bodies are scanned too, which only errs on the safe
// side.
func repeatable(ctx *analysis.Context, e syntax.Expr) bool {
	ok := true
	// A rejected node only ends the walk below itself; later siblings are
	// still visited, so the verdict may only ever go from true to false.
	syntax.Inspect(e, func(n syntax.Node) bool {
		if !ok {
			return false
		}
		switch x := n.(type) {
		case *syntax.FuncCall:
			call, _ := semanticquery.GlobalCall(ctx, x, pureFunctions...)
			ok = call != nil
		case *syntax.MethodCall, *syntax.StaticCall, *syntax.New, *syntax.Clone, *syntax.Assign,
			*syntax.IncDec, *syntax.Include, *syntax.Eval, *syntax.Exit, *syntax.Print,
			*syntax.Throw, *syntax.Yield, *syntax.YieldFrom:
			ok = false
		case *syntax.Binary:
			ok = x.Op.Kind != syntax.TPipe
		case *syntax.InterpolatedString:
			ok = !x.Backtick
		}
		return ok
	})
	return ok
}
