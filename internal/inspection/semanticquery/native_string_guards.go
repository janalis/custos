package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// NativeStringNonNullGuard recognizes tests that null, coerced to an empty
// string by a native string operation, cannot satisfy. The tested binding must
// remain unchanged until the read, including across a loop back-edge.
func NativeStringNonNullGuard(ctx *analysis.Context, input syntax.Expr) bool {
	accepts := func(condition syntax.Expr, truth bool) bool {
		if call, ok := nativeStringGuardExpr(condition).(*syntax.FuncCall); ok && truth {
			if _, bound := ctx.Flow().BoundArguments(call); !bound {
				return false
			}
			switch NativeBuiltinName(ctx, call) {
			case "str_starts_with", "str_ends_with", "str_contains":
				needle, known := astquery.QuotedStringValue(CallArgument(call.Args, 1, "needle"))
				return known && needle != "" && nativeStringGuardSameValue(ctx, input, CallArgument(call.Args, 0, "haystack"))
			}
		}
		b, ok := syntax.UnwrapParens(condition).(*syntax.Binary)
		if !ok || b.Op.Kind != syntax.TIsIdentical && b.Op.Kind != syntax.TIsNotIdentical {
			return false
		}
		equal := truth == (b.Op.Kind == syntax.TIsIdentical)
		for _, pair := range [][2]syntax.Expr{{b.Left, b.Right}, {b.Right, b.Left}} {
			if call, ok := nativeStringGuardExpr(pair[0]).(*syntax.FuncCall); ok && !equal && NativeBuiltin(ctx, call, "strpos") {
				constant, ok := syntax.UnwrapParens(pair[1]).(*syntax.ConstFetch)
				needle, known := astquery.QuotedStringValue(CallArgument(call.Args, 1, "needle"))
				_, bound := ctx.Flow().BoundArguments(call)
				if bound && ok && constant.Name != nil && strings.EqualFold(strings.TrimPrefix(constant.Name.Value, `\`), "false") && known && needle != "" && nativeStringGuardSameValue(ctx, input, CallArgument(call.Args, 0, "haystack")) {
					return true
				}
			}
			if offset, ok := syntax.UnwrapParens(pair[0]).(*syntax.ArrayDimFetch); ok && equal {
				text, known := astquery.QuotedStringValue(syntax.UnwrapParens(pair[1]))
				if known && text != "" && nativeStringGuardSameValue(ctx, input, offset.Var) {
					return true
				}
			}
		}
		return false
	}
	if nativePathGuard(ctx, input, accepts) {
		return true
	}
	if NativeLabelled(ctx, input) {
		return false
	}
	budget := 512
	for p := input.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		statement, ok := p.(syntax.Stmt)
		if !ok {
			continue
		}
		list, i, _ := astquery.StmtList(ctx.File, statement)
		for _, previous := range list[:max(i, 0)] {
			budget--
			if budget < 0 {
				return false
			}
			guard, ok := previous.(*syntax.If)
			if ok && guard.Else == nil && len(guard.ElseIfs) == 0 && syntax.Terminates(guard.Body) && nativeGuardHolds(guard.Cond, false, accepts) {
				return true
			}
		}
	}
	return false
}

func nativeStringGuardExpr(e syntax.Expr) syntax.Expr {
	e = syntax.UnwrapParens(e)
	if assignment, ok := e.(*syntax.Assign); ok && assignment.Op.Kind == syntax.TEqual && !assignment.ByRef {
		return syntax.UnwrapParens(assignment.Value)
	}
	return e
}

func nativeStringGuardSameValue(ctx *analysis.Context, input, tested syntax.Expr) bool {
	if !astquery.Equivalent(ctx.File, input, tested) || NativeLabelled(ctx, input) {
		return false
	}
	variable, ok := syntax.UnwrapParens(input).(*syntax.Variable)
	if !ok || variable.NameExpr != nil {
		return false
	}
	scope := syntax.EnclosingVariableScope(input)
	writes := nativeContractScope(ctx, scope).localWrites(ctx, scope)
	for parent := input.Parent(); parent != nil && parent != scope; parent = parent.Parent() {
		switch parent.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			if !parent.Span().Contains(tested.Span()) && writes.written(variable.Name, parent.Span().Start, parent.Span().End) {
				return false
			}
		}
	}
	if nativeGuardSameValue(ctx, input, tested) {
		return true
	}
	return NativeUnreplaced(ctx, input, tested) && !nativeStringRewritten(writes, variable.Name, tested.Span().End, input)
}

func nativeStringRewritten(writes *nativeLocalWrites, name string, after uint32, input syntax.Expr) bool {
	if !writes.known {
		return true
	}
	before := input.Span().Start
	for _, write := range writes.writes {
		if write.name != name && write.name != "" || write.at <= after || write.at >= before {
			continue
		}
		own := false
		if write.plain {
			for parent := input.Parent(); parent != nil && !syntax.IsVariableScope(parent); parent = parent.Parent() {
				if assignment, ok := parent.(*syntax.Assign); ok && assignment.Span().Start == write.at && assignment.Value.Span().Contains(input.Span()) {
					own = true
					break
				}
			}
		}
		if !own {
			return true
		}
	}
	return writes.reached(name, after, before)
}
