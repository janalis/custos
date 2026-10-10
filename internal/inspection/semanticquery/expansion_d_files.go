package semanticquery

import (
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// ExpansionDLocalEvents returns bounded lexical effects, excluding nested scopes.
func ExpansionDLocalEvents(ctx *analysis.Context, at syntax.Node) []syntax.Node {
	return nativeContractEvents(ctx, syntax.EnclosingVariableScope(at))
}

// ExpansionDPathCall identifies a builtin consuming its first argument as a path.
func ExpansionDPathCall(ctx *analysis.Context, c *syntax.FuncCall) bool {
	switch NativeBuiltinName(ctx, c) {
	case "file_get_contents", "file_put_contents", "fopen", "file", "is_file", "is_dir", "file_exists", "unlink", "mkdir", "rmdir", "scandir", "opendir":
		return true
	}
	return false
}

// ExpansionDFinalOutcome recognizes an explicit terminal result, allowing throws
// only when inspecting a failure branch.
func ExpansionDFinalOutcome(ctx *analysis.Context, body syntax.Stmt, success bool) bool {
	if b, ok := body.(*syntax.Block); ok {
		if len(b.Stmts) == 0 {
			return false
		}
		body = b.Stmts[len(b.Stmts)-1]
	}
	if !NativeCallbackReachable(body) {
		return false
	}
	if st, ok := body.(*syntax.ExprStmt); ok {
		if _, thrown := syntax.UnwrapParens(st.Expr).(*syntax.Throw); thrown {
			return !success
		}
	}
	ret, ok := body.(*syntax.Return)
	if !ok {
		return false
	}
	c, ok := syntax.UnwrapParens(ret.Expr).(*syntax.ConstFetch)
	if !ok {
		return false
	}
	name := GlobalConstName(ctx, c)
	if success {
		return name == "true"
	}
	return name == "false"
}

// ExpansionDCallbackTerminates proves termination through literal callback
// conditions as well as the syntax-level terminal statements. Depth is bounded.
func ExpansionDCallbackTerminates(ctx *analysis.Context, body syntax.Stmt) bool {
	return expansionDCallbackTerminates(ctx, body, 64)
}

func expansionDCallbackTerminates(ctx *analysis.Context, body syntax.Stmt, budget int) bool {
	if budget == 0 {
		return false
	}
	if syntax.Terminates(body) {
		return true
	}
	switch b := body.(type) {
	case *syntax.Block:
		for _, stmt := range b.Stmts {
			if expansionDCallbackTerminates(ctx, stmt, budget-1) {
				return true
			}
		}
	case *syntax.If:
		truth, known := NativeTruth(ctx, b.Cond)
		if known && len(b.ElseIfs) == 0 {
			if truth {
				return expansionDCallbackTerminates(ctx, b.Body, budget-1)
			}
			if b.Else != nil {
				return expansionDCallbackTerminates(ctx, b.Else.Body, budget-1)
			}
		}
	}
	return false
}
