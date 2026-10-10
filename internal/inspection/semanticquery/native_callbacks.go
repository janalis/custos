package semanticquery

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/php/syntax"
)

// NativeCallback resolves an inline or same-file named callback without
// executing its body. Unknown dynamic callbacks remain unknown.
func NativeCallback(ctx *analysis.Context, e syntax.Expr) ([]*syntax.Param, syntax.Node) {
	v := NativeValue(ctx, e)
	params, body := astquery.ScopeParts(v)
	if body != nil {
		return params, body
	}
	if name, ok := NativeString(ctx, e); ok {
		if fn := ctx.Index().Function(name, ctx.PHP); fn != nil {
			if declaration := FunctionDecl(ctx.File, fn); declaration != nil {
				return astquery.ScopeParts(declaration)
			}
		}
	}
	return nil, nil
}

// NativeReturns collects explicit returns of one callback, excluding nested
// scopes. A bare return or generator prevents a complete value-only contract.
func NativeReturns(body syntax.Node) ([]syntax.Expr, bool) {
	if body == nil {
		return nil, false
	}
	if e, ok := body.(syntax.Expr); ok {
		return []syntax.Expr{e}, true
	}
	block, ok := body.(*syntax.Block)
	if !ok || !syntax.Terminates(block) {
		return nil, false
	}
	var values []syntax.Expr
	complete := true
	syntax.Inspect(body, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.Function, *syntax.Closure, *syntax.ArrowFunction, *syntax.ClassLike:
			return false
		case *syntax.Return:
			if n.Expr == nil {
				complete = false
			} else {
				values = append(values, n.Expr)
			}
		case *syntax.Yield, *syntax.YieldFrom:
			complete = false
		}
		return true
	})
	return values, complete && len(values) > 0
}

// NativeFunctionBody returns a same-file resolved named function body.
func NativeFunctionBody(ctx *analysis.Context, call *syntax.FuncCall) *syntax.Block {
	fn := ctx.Types().ResolveFunction(call)
	decl := FunctionDecl(ctx.File, fn)
	if decl == nil {
		return nil
	}
	return decl.Body
}

// NativeGeneratorBody recognizes yields owned by this body, not nested scopes.
func NativeGeneratorBody(body *syntax.Block) bool {
	if body == nil {
		return false
	}
	found := false
	syntax.Inspect(body, func(n syntax.Node) bool {
		switch n.(type) {
		case *syntax.Function, *syntax.Closure, *syntax.ArrowFunction, *syntax.ClassLike:
			return false
		case *syntax.Yield, *syntax.YieldFrom:
			found = true
		}
		return !found
	})
	return found
}

// NativeCallbackReachable excludes callback statements after termination and
// bodies whose literal condition proves they cannot execute. Unknown branches
// remain possible; this is a reachability filter, not execution prediction.
func NativeCallbackReachable(n syntax.Node) bool {
	scope := syntax.EnclosingVariableScope(n)
	if !flowquery.Reachable(n, scope) {
		return false
	}
	for parent := n.Parent(); parent != nil && parent != scope; parent = parent.Parent() {
		if branch, ok := parent.(*syntax.If); ok {
			truth, known := astquery.BoolConst(syntax.UnwrapParens(branch.Cond))
			if known && !truth && branch.Body.Span().Contains(n.Span()) {
				return false
			}
			if known && truth && branch.Else != nil && branch.Else.Body.Span().Contains(n.Span()) {
				return false
			}
		}
	}
	return true
}
