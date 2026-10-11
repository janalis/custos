package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/php/syntax"
)

// NativeLengthGuard proves that an enclosing true branch requires an exact
// byte length. Other validation shapes stay unknown rather than guessed safe.
func NativeLengthGuard(ctx *analysis.Context, input syntax.Expr, length int64) bool {
	if input == nil {
		return false
	}
	if n, known := ctx.Flow().ExactLength(input); known && n == length {
		return true
	}
	return nativePathGuard(ctx, input, func(condition syntax.Expr, truth bool) bool {
		b, ok := syntax.UnwrapParens(condition).(*syntax.Binary)
		if !ok || !((truth && b.Op.Kind == syntax.TIsIdentical) || (!truth && b.Op.Kind == syntax.TIsNotIdentical)) {
			return false
		}
		for _, pair := range [][2]syntax.Expr{{b.Left, b.Right}, {b.Right, b.Left}} {
			call, ok := syntax.UnwrapParens(pair[0]).(*syntax.FuncCall)
			if !ok || !NativeBuiltin(ctx, call, "strlen") {
				continue
			}
			n, known := NativeInt(ctx, pair[1])
			return known && n == length && nativeGuardSameValue(ctx, CallArgument(call.Args, 0, "string"), input)
		}
		return false
	})
}

// NativeSentinelGuard proves that the null/false failure value cannot reach
// the input: the flow facts of the tested variable exclude it, or a strict
// comparison of the same value guards the current path.
func NativeSentinelGuard(ctx *analysis.Context, input syntax.Expr, sentinel string) bool {
	if input == nil {
		return false
	}
	if ctx.Flow().Excludes(input, sentinel) {
		return true
	}
	return nativePathGuard(ctx, input, func(condition syntax.Expr, truth bool) bool {
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

// nativePathGuard finds a condition whose outcome is fixed wherever input
// runs: the test of an enclosing if/elseif/else arm, the deciding left operand
// of a short-circuit operator, a ternary condition, or a terminating `if`
// written directly before an enclosing statement. Negations, true conjunctions
// and false disjunctions are decomposed before accepts sees an operand.
func nativePathGuard(ctx *analysis.Context, input syntax.Expr, accepts func(syntax.Expr, bool) bool) bool {
	var child syntax.Node = input
	for p := input.Parent(); p != nil; child, p = p, p.Parent() {
		if syntax.IsVariableScope(p) {
			break
		}
		switch x := p.(type) {
		case *syntax.If:
			if x.Body == child && nativeGuardHolds(x.Cond, true, accepts) {
				return true
			}
		case *syntax.ElseIf:
			if x.Body == child && nativeGuardHolds(x.Cond, true, accepts) || nativeGuardEarlierArms(x, accepts) {
				return true
			}
		case *syntax.Else:
			if nativeGuardEarlierArms(x, accepts) {
				return true
			}
		case *syntax.Binary:
			if x.Right == child {
				switch x.Op.Kind {
				case syntax.TBooleanAnd, syntax.TAnd:
					if nativeGuardHolds(x.Left, true, accepts) {
						return true
					}
				case syntax.TBooleanOr, syntax.TOr:
					if nativeGuardHolds(x.Left, false, accepts) {
						return true
					}
				}
			}
		case *syntax.Ternary:
			if x.Then != nil && x.Then == child && nativeGuardHolds(x.Cond, true, accepts) || x.Else == child && nativeGuardHolds(x.Cond, false, accepts) {
				return true
			}
		}
		if st, ok := p.(syntax.Stmt); ok {
			previous, exists := astquery.PrevStmt(ctx.File, st)
			if guard, ok := previous.(*syntax.If); exists && ok && guard.Else == nil && len(guard.ElseIfs) == 0 && syntax.Terminates(guard.Body) && nativeGuardHolds(guard.Cond, false, accepts) {
				return true
			}
		}
	}
	return false
}

// nativeGuardEarlierArms handles an elseif or else arm: every condition
// written before it was false when the arm runs.
func nativeGuardEarlierArms(arm syntax.Node, accepts func(syntax.Expr, bool) bool) bool {
	owner := arm.Parent().(*syntax.If)
	if nativeGuardHolds(owner.Cond, false, accepts) {
		return true
	}
	for _, other := range owner.ElseIfs {
		if syntax.Node(other) == arm {
			break
		}
		if nativeGuardHolds(other.Cond, false, accepts) {
			return true
		}
	}
	return false
}

// nativeGuardHolds decomposes a condition with a known outcome: a negation
// flips it, a true conjunction and a false disjunction fix both operands.
func nativeGuardHolds(cond syntax.Expr, truth bool, accepts func(syntax.Expr, bool) bool) bool {
	switch c := syntax.UnwrapParens(cond).(type) {
	case *syntax.Unary:
		if c.Op.Kind == syntax.TExclaim {
			return nativeGuardHolds(c.Expr, !truth, accepts)
		}
	case *syntax.Binary:
		switch c.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd:
			if truth {
				return nativeGuardHolds(c.Left, true, accepts) || nativeGuardHolds(c.Right, true, accepts)
			}
			return false
		case syntax.TBooleanOr, syntax.TOr:
			if !truth {
				return nativeGuardHolds(c.Left, false, accepts) || nativeGuardHolds(c.Right, false, accepts)
			}
			return false
		}
	}
	return accepts(cond, truth)
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
	// Without a reaching producer only a plain read is the same value twice,
	// and a variable only while nothing writes it between the two reads.
	if flowquery.MayHaveSideEffects(a) {
		return false
	}
	variable, ok := syntax.UnwrapParens(a).(*syntax.Variable)
	if !ok || variable.NameExpr != nil {
		return true
	}
	scope := syntax.EnclosingVariableScope(a)
	after, before := a.Span().End, b.Span().Start
	if after > before {
		after, before = b.Span().End, a.Span().Start
	}
	return !nativeContractScope(ctx, scope).localWrites(ctx, scope).written(variable.Name, after, before)
}
