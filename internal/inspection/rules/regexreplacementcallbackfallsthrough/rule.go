// Package regexreplacementcallbackfallsthrough implements the native RegexReplacementCallbackFallsThrough inspection.
package regexreplacementcallbackfallsthrough

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Return the replacement from the callback."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "RegexReplacementCallbackFallsThrough" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "preg_replace_callback") {
		return
	}
	value := semanticquery.NativeValue(ctx, semanticquery.CallArgument(c.Args, 1, "callback"))
	closure, ok := value.(*syntax.Closure)
	if !ok || closure.Body == nil || semanticquery.NativeGeneratorBody(closure.Body) {
		return
	}
	if completes(ctx, closure.Body) && supported(closure.Body) {
		ctx.ReportNode(c, message)
	}
}

// completes reports whether control provably reaches the end of body. Literal
// conditions select their branch; a call that never returns, or whose callee
// cannot be resolved, ends the path.
func completes(ctx *analysis.Context, body syntax.Stmt) bool {
	if syntax.Terminates(body) {
		return false
	}
	switch b := body.(type) {
	case *syntax.Block:
		for _, stmt := range b.Stmts {
			if !completes(ctx, stmt) {
				return false
			}
		}
	case *syntax.ExprStmt:
		return returns(ctx, syntax.UnwrapParens(b.Expr))
	case *syntax.If:
		if truth, known := semanticquery.NativeTruth(ctx, b.Cond); known && len(b.ElseIfs) == 0 {
			if truth {
				return completes(ctx, b.Body)
			}
			return b.Else == nil || completes(ctx, b.Else.Body)
		}
		if b.Else == nil || completes(ctx, b.Body) || completes(ctx, b.Else.Body) {
			return true
		}
		for _, branch := range b.ElseIfs {
			if completes(ctx, branch.Body) {
				return true
			}
		}
		return false
	}
	return true
}

// returns reports whether evaluating a statement expression provably hands
// control back: its type is not `never`, and a function callee or a method
// receiver is known, since an unknown declaration may be declared never.
func returns(ctx *analysis.Context, e syntax.Expr) bool {
	if ctx.TypeOf(e).OnlyOf("never") {
		return false
	}
	switch c := e.(type) {
	case *syntax.FuncCall:
		return ctx.Types().ResolveFunction(c) != nil
	case *syntax.MethodCall:
		return !ctx.TypeOf(c.Var).IsUnknown()
	}
	return true
}

func supported(body syntax.Node) bool {
	ok := true
	syntax.Inspect(body, func(n syntax.Node) bool {
		if syntax.IsVariableScope(n) || n.Kind() == syntax.KClassLike {
			return false
		}
		switch n.(type) {
		case *syntax.While, *syntax.For, *syntax.Foreach, *syntax.DoWhile, *syntax.Try, *syntax.Switch, *syntax.Goto:
			ok = false
		}
		return ok
	})
	return ok
}
