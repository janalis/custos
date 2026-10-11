package untrustedheadervalue

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/flow"
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
		if !origin.Complete || origin.Expr == nil || aliased(ctx, x) {
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

// guarded finds a line-break rejection of the value read by use: an earlier
// sibling `if` that leaves on a positive test, or an enclosing branch entered
// only on a negative one. origin is the single assignment reaching use; the
// tested expression must still carry that same assignment.
func guarded(ctx *analysis.Context, use syntax.Node, origin syntax.Expr) bool {
	for node := use; node != nil && !syntax.IsVariableScope(node); node = node.Parent() {
		switch parent := node.Parent().(type) {
		case *syntax.If:
			if node == parent.Body && clean(ctx, parent.Cond, origin, true) {
				return true
			}
		case *syntax.Else:
			if owner, ok := parent.Parent().(*syntax.If); ok && clean(ctx, owner.Cond, origin, false) {
				return true
			}
		}
		list, ok := syntax.StmtListOf(node.Parent())
		if node.Parent() == nil {
			list, ok = ctx.File.Stmts, true
		}
		if !ok {
			continue
		}
		for _, earlier := range list {
			if earlier.Span().Start >= node.Span().Start {
				break
			}
			if guard, ok := earlier.(*syntax.If); ok && syntax.Terminates(guard.Body) && clean(ctx, guard.Cond, origin, false) {
				return true
			}
		}
	}
	return false
}

// clean reports whether cond evaluating to truth proves the value assigned by
// origin holds neither CR nor LF.
func clean(ctx *analysis.Context, cond, origin syntax.Expr, truth bool) bool {
	switch x := syntax.UnwrapParens(cond).(type) {
	case *syntax.FuncCall:
		return !truth && lineBreakTest(ctx, x, origin) != ""
	case *syntax.Unary:
		return x.Op.Kind == syntax.TExclaim && clean(ctx, x.Expr, origin, !truth)
	case *syntax.Binary:
		equal := false
		switch x.Op.Kind {
		case syntax.TBooleanOr, syntax.TOr:
			return !truth && (clean(ctx, x.Left, origin, false) || clean(ctx, x.Right, origin, false))
		case syntax.TBooleanAnd, syntax.TAnd:
			return truth && (clean(ctx, x.Left, origin, true) || clean(ctx, x.Right, origin, true))
		case syntax.TGreater:
			zero, ok := syntax.UnwrapParens(x.Right).(*syntax.Literal)
			return !truth && ok && zero.Raw == "0" && lineBreakTest(ctx, x.Left, origin) == "preg_match"
		case syntax.TIsIdentical, syntax.TIsEqual:
			equal = true
		case syntax.TIsNotIdentical, syntax.TIsNotEqual:
		default:
			return false
		}
		test, operand := x.Left, x.Right
		if _, ok := syntax.UnwrapParens(test).(*syntax.FuncCall); !ok {
			test, operand = operand, test
		}
		// With a line break present the comparison yields found == equal.
		found, known := false, false
		switch o := syntax.UnwrapParens(operand).(type) {
		case *syntax.Literal: // preg_match returns 1 for a matching subject
			found, known = o.Raw == "1", (o.Raw == "0" || o.Raw == "1") && lineBreakTest(ctx, test, origin) == "preg_match"
		case *syntax.ConstFetch:
			known = strings.EqualFold(strings.TrimPrefix(o.Name.Value, `\`), "false") && lineBreakTest(ctx, test, origin) != ""
		}
		return known && (found == equal) != truth
	}
	return false
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
