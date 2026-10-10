package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// NativeLengthGuard proves that an enclosing true branch requires an exact
// byte length. Other validation shapes stay unknown rather than guessed safe.
func NativeLengthGuard(ctx *analysis.Context, input syntax.Expr, length int64) bool {
	return nativeGuard(ctx, input, func(condition syntax.Expr, truth bool) bool {
		b, ok := syntax.UnwrapParens(condition).(*syntax.Binary)
		if !ok || !((truth && b.Op.Kind == syntax.TIsIdentical) || (!truth && b.Op.Kind == syntax.TIsNotIdentical)) {
			return false
		}
		call, ok := syntax.UnwrapParens(b.Left).(*syntax.FuncCall)
		if !ok || !NativeBuiltin(ctx, call, "strlen") {
			return false
		}
		n, known := NativeInt(ctx, b.Right)
		return known && n == length && nativeGuardSameValue(ctx, CallArgument(call.Args, 0, "string"), input)
	})
}

// NativeSentinelGuard proves that a strict null/false guard excludes a failure
// value on the current path. It never treats general truthiness as validation.
func NativeSentinelGuard(ctx *analysis.Context, input syntax.Expr, sentinel string) bool {
	return nativeGuard(ctx, input, func(condition syntax.Expr, truth bool) bool {
		b, ok := syntax.UnwrapParens(condition).(*syntax.Binary)
		if !ok || !((truth && b.Op.Kind == syntax.TIsNotIdentical) || (!truth && b.Op.Kind == syntax.TIsIdentical)) {
			return false
		}
		for _, pair := range [][2]syntax.Expr{{b.Left, b.Right}, {b.Right, b.Left}} {
			constant, ok := syntax.UnwrapParens(pair[1]).(*syntax.ConstFetch)
			if ok && constant.Name != nil && strings.EqualFold(strings.TrimPrefix(constant.Name.Value, `\`), sentinel) && nativeGuardSameValue(ctx, input, pair[0]) {
				return true
			}
		}
		return false
	})
}

func nativeGuard(ctx *analysis.Context, input syntax.Expr, accepts func(syntax.Expr, bool) bool) bool {
	if input == nil {
		return false
	}
	for p := input.Parent(); p != nil; p = p.Parent() {
		if syntax.IsVariableScope(p) {
			break
		}
		branch, ok := p.(*syntax.If)
		if ok && branch.Body != nil && branch.Body.Span().Contains(input.Span()) && accepts(branch.Cond, true) {
			return true
		}
		if ok && branch.Else != nil && branch.Else.Body.Span().Contains(input.Span()) && accepts(branch.Cond, false) {
			return true
		}
		if st, ok := p.(syntax.Stmt); ok {
			previous, exists := astquery.PrevStmt(ctx.File, st)
			if guard, ok := previous.(*syntax.If); exists && ok && guard.Else == nil && len(guard.ElseIfs) == 0 && syntax.Terminates(guard.Body) && accepts(guard.Cond, false) {
				return true
			}
		}
	}
	return false
}

// NativeIntegerAllowsZero proves the validator options neither exclude zero
// nor change the failure sentinel to null.
func NativeIntegerAllowsZero(ctx *analysis.Context, options syntax.Expr) bool {
	if options == nil {
		return true
	}
	if n, known := NativeInt(ctx, options); known {
		return n&134217728 == 0 // FILTER_NULL_ON_FAILURE
	}
	entries, known := NativeArrayEntries(ctx, options)
	if !known {
		return false
	}
	if flags := entries["s:flags"]; flags != nil {
		present, known := NativeFlag(ctx, flags, 134217728)
		if !known || present {
			return false
		}
	}
	if bounds := entries["s:options"]; bounds != nil {
		values, known := NativeArrayEntries(ctx, bounds)
		if !known {
			return false
		}
		for _, key := range []string{"min_range", "max_range"} {
			if value := values["s:"+key]; value != nil {
				n, known := NativeInt(ctx, value)
				if !known || (key == "min_range" && n > 0) || (key == "max_range" && n < 0) {
					return false
				}
			}
		}
	}
	return true
}

// NativeCurlDefault proves a local handle retains an option's disabled default.
// Any escape, ambiguous option, bulk update or write callback invalidates proof.
func NativeCurlDefault(ctx *analysis.Context, at syntax.Node, handle syntax.Expr, option string) bool {
	id := ctx.Flow().Value(handle).Identity
	if id == 0 {
		return false
	}
	for _, call := range ctx.Flow().Calls(syntax.EnclosingVariableScope(at)) {
		if call.Node.Span().Start >= at.Span().Start {
			continue
		}
		uses := false
		for _, arg := range call.Arguments {
			if arg.Identity == id {
				uses = true
			}
		}
		if !uses {
			continue
		}
		fc, ok := call.Node.(*syntax.FuncCall)
		if !ok {
			return false
		}
		switch NativeBuiltinName(ctx, fc) {
		case "curl_setopt":
			opt := CallArgument(fc.Args, 1, "option")
			constant, ok := syntax.UnwrapParens(opt).(*syntax.ConstFetch)
			if !ok || constant.Name == nil {
				return false
			}
			name := strings.TrimPrefix(constant.Name.Value, `\`)
			if name == "CURLOPT_WRITEFUNCTION" {
				return false
			}
			if name == option {
				truth, known := NativeTruth(ctx, CallArgument(fc.Args, 2, "value"))
				if !known || truth {
					return false
				}
			}
		case "curl_init":
		default:
			return false
		}
	}
	return true
}

// nativeGuardSameValue keeps syntactic guards tied to the reaching producer.
// A repeated call or reassignment creates a new value even when its text agrees.
func nativeGuardSameValue(ctx *analysis.Context, a, b syntax.Expr) bool {
	if !astquery.Equivalent(ctx.File, a, b) {
		return false
	}
	av, bv := ctx.Flow().Value(a), ctx.Flow().Value(b)
	if av.Expr != nil || bv.Expr != nil {
		return av.Expr == bv.Expr && av.Invalidated == bv.Invalidated
	}
	return true
}
