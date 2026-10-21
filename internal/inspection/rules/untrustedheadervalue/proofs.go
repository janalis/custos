package untrustedheadervalue

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/types"
)

// maxDepth bounds the structural walk through operands and definitions.
const maxDepth = 32

// scalarBuiltins return an integer or a boolean (or false/null on failure),
// whatever bytes their arguments hold.
var scalarBuiltins = map[string]bool{
	"strlen": true, "mb_strlen": true, "iconv_strlen": true, "count": true, "sizeof": true,
	"crc32": true, "ord": true, "strcmp": true, "strcasecmp": true, "strncmp": true,
	"strncasecmp": true, "strnatcmp": true, "strnatcasecmp": true, "strpos": true,
	"stripos": true, "strrpos": true, "strripos": true, "mb_strpos": true, "mb_stripos": true,
	"mb_strrpos": true, "mb_strripos": true, "substr_count": true, "mb_substr_count": true,
	"levenshtein": true, "similar_text": true, "in_array": true, "array_key_exists": true,
	"key_exists": true, "str_contains": true, "str_starts_with": true, "str_ends_with": true,
}

// inert proves that no request-controlled CR or LF byte can appear in the
// string conversion of e. Every part the structure does not settle must be
// completely analysed and free of header taint.
func inert(ctx *analysis.Context, e syntax.Expr, depth int) bool {
	e = syntax.UnwrapParens(e)
	if depth == 0 {
		return false
	}
	if v := ctx.Flow().Value(e); v.Complete && !ctx.Flow().Tainted(e, flow.Header) {
		return true
	}
	switch x := e.(type) {
	case *syntax.InterpolatedString:
		for _, part := range x.Parts {
			if !inert(ctx, part, depth-1) {
				return false
			}
		}
		return true
	case *syntax.Binary:
		switch x.Op.Kind {
		case syntax.TDot:
			return inert(ctx, x.Left, depth-1) && inert(ctx, x.Right, depth-1)
		case syntax.TPlus: // array + array is a union keeping its elements
			return scalar(ctx, x.Left) || scalar(ctx, x.Right)
		}
	case *syntax.Unary:
		return x.Op.Kind == syntax.TPlus || x.Op.Kind == syntax.TMinus
	case *syntax.Ternary:
		first := x.Then
		if first == nil {
			first = x.Cond
		}
		return inert(ctx, first, depth-1) && inert(ctx, x.Else, depth-1)
	case *syntax.Match: // the subject only selects an arm
		for _, arm := range x.Arms {
			if _, throws := syntax.UnwrapParens(arm.Body).(*syntax.Throw); !throws && !inert(ctx, arm.Body, depth-1) {
				return false
			}
		}
		return true
	case *syntax.FuncCall:
		return scalarCall(ctx, x)
	case *syntax.Variable:
		origin := ctx.Flow().Value(x)
		if !origin.Complete || origin.Expr == nil || semanticquery.RebindableVariable(ctx, x) {
			return false
		}
		return guarded(ctx, x, origin.Expr) || inert(ctx, origin.Expr, depth-1)
	}
	return false
}

// scalar recognizes an operand that cannot be an array, so `+` is arithmetic.
func scalar(ctx *analysis.Context, e syntax.Expr) bool {
	switch x := syntax.UnwrapParens(e).(type) {
	case *syntax.Literal:
		return true
	case *syntax.Unary:
		switch x.Op.Kind {
		case syntax.TIntCast, syntax.TDoubleCast, syntax.TBoolCast, syntax.TStringCast:
			return true
		}
	case *syntax.FuncCall:
		return scalarCall(ctx, x)
	}
	return false
}

func scalarCall(ctx *analysis.Context, call *syntax.FuncCall) bool {
	name := semanticquery.NativeBuiltinName(ctx, call)
	return scalarBuiltins[name] || strings.HasPrefix(name, "ctype_") || strings.HasPrefix(name, "is_")
}

// guarded finds a line-break rejection of the value read by use: a
// preg_match or strpbrk test of that very value (origin is the single
// assignment reaching use, and the tested expression must still carry it)
// whose result the conditions guarding use prove negative.
func guarded(ctx *analysis.Context, use, origin syntax.Expr) bool {
	if semanticquery.NativeLabelled(ctx, use) {
		return false // a jump may land between the test and the use
	}
	return semanticquery.NativeEarlierCallProves(ctx, use, func(x syntax.Expr) (types.Type, func(types.Type) bool) {
		switch lineBreakTest(ctx, x, origin) {
		case "preg_match":
			// 1 when a line break matched, 0 when none did, false on error.
			return types.Union(types.Ints(types.IntRange{Lo: 0, Hi: 1}), types.Of("false")), func(t types.Type) bool { return !t.HasInt(1) }
		case "strpbrk":
			// The rest of the subject from the first line break (never '' nor
			// '0'), or false.
			found := types.TrackedStrings().StringsWithout("").StringsWithout("0")
			return types.Union(found, types.Of("false")), func(t types.Type) bool { return t.OnlyOf("false") }
		}
		return types.Unknown, nil
	})
}

// lineBreakTest names the builtin when e is a two-argument preg_match or
// strpbrk call that is truthy exactly when the value assigned by origin
// contains CR or LF. Other shapes yield "".
func lineBreakTest(ctx *analysis.Context, e, origin syntax.Expr) string {
	call, ok := syntax.UnwrapParens(e).(*syntax.FuncCall)
	if !ok || call.Args == nil || len(call.Args.Args) != 2 {
		return ""
	}
	name := semanticquery.NativeBuiltinName(ctx, call)
	var set, subject syntax.Expr
	switch name {
	case "preg_match":
		set, subject = semanticquery.CallArgument(call.Args, 0, "pattern"), semanticquery.CallArgument(call.Args, 1, "subject")
	case "strpbrk":
		subject, set = semanticquery.CallArgument(call.Args, 0, "string"), semanticquery.CallArgument(call.Args, 1, "characters")
	default:
		return ""
	}
	text, known := astquery.QuotedStringValue(syntax.UnwrapParens(set))
	tested := ctx.Flow().Value(subject)
	if !known || !tested.Complete || tested.Expr != origin {
		return ""
	}
	if name == "preg_match" {
		// Only a delimited, modifier-free class of the two bytes cannot fail.
		n := len(text)
		if n < 4 || text[0] != text[n-1] || text[1] != '[' || text[n-2] != ']' || !strings.Contains("/#~", text[:1]) {
			return ""
		}
		text = strings.NewReplacer(`\r`, "\r", `\n`, "\n").Replace(text[2 : n-2])
		if strings.Trim(text, "\r\n") != "" {
			return ""
		}
	}
	if !strings.Contains(text, "\r") || !strings.Contains(text, "\n") {
		return ""
	}
	return name
}
