package semanticquery

import (
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// ExpansionDLocalEvents returns bounded lexical effects, excluding nested scopes.
func ExpansionDLocalEvents(ctx *analysis.Context, at syntax.Node) []syntax.Node {
	return ContractEvents(ctx, syntax.EnclosingVariableScope(at))
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
	if !NativeCallbackReachable(ctx, body) {
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
