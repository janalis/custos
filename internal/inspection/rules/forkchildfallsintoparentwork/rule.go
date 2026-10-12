// Package forkchildfallsintoparentwork implements the native ForkChildFallsIntoParentWork inspection.
package forkchildfallsintoparentwork

import (
	"math"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

const message = "Terminate or separate the fork child branch."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ForkChildFallsIntoParentWork" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KIf} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	branch := n.(*syntax.If)
	if branch.Else != nil || len(branch.ElseIfs) > 0 {
		return
	}
	cond, ok := syntax.UnwrapParens(branch.Cond).(*syntax.Binary)
	if !ok || cond.Op.Kind != syntax.TIsIdentical {
		return
	}
	zero, known := semanticquery.NativeInt(ctx, cond.Right)
	if !known || zero != 0 {
		return
	}
	fork, ok := semanticquery.NativeValue(ctx, cond.Left).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, fork, "pcntl_fork") {
		return
	}
	body, ok := branch.Body.(*syntax.Block)
	if !ok || len(body.Stmts) == 0 {
		return
	}
	if leaves(body) || unknownTail(ctx, body) {
		return
	}
	if _, ok := astquery.NextStmt(ctx.File, branch); ok {
		ctx.ReportNode(cond, message)
	}
}

// leaves reports whether control can never reach the statement after body:
// the block ends abruptly and no break, continue or goto carries the child
// out of it into code the parent also runs.
func leaves(body *syntax.Block) bool {
	return syntax.Terminates(body) && !jumpsOut(body, 0)
}

// jumpsOut reports whether n holds a goto, or a break/continue that reaches
// past the loops and switches opened inside the child block: one for more
// levels than depth counts. A level that is not a literal is assumed to.
func jumpsOut(n syntax.Node, depth int64) bool {
	switch x := n.(type) {
	case *syntax.Function, *syntax.Closure, *syntax.ArrowFunction, *syntax.ClassLike:
		return false
	case *syntax.Goto:
		return true
	case *syntax.Break:
		return jumpLevels(x.Num) > depth
	case *syntax.Continue:
		return jumpLevels(x.Num) > depth
	case *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach, *syntax.Switch:
		depth++
	}
	out := false
	syntax.Children(n, func(c syntax.Node) { out = out || jumpsOut(c, depth) })
	return out
}

// jumpLevels is the number of levels a break or continue leaves: one without
// an operand, the literal operand, and more than any nesting otherwise.
func jumpLevels(operand syntax.Expr) int64 {
	if operand == nil {
		return 1
	}
	if literal, ok := syntax.UnwrapParens(operand).(*syntax.Literal); ok {
		if levels, known := astquery.ParseIntLiteral(literal.Raw); known && levels > 0 {
			return levels
		}
	}
	return math.MaxInt64
}

// unknownTail reports whether the last effective statement of body is a call
// whose termination cannot be summarised (no resolvable declaration or body) or
// whose declaration never returns. Such a child may well end inside the callee.
func unknownTail(ctx *analysis.Context, body *syntax.Block) bool {
	var tail *syntax.ExprStmt
	for i := len(body.Stmts) - 1; i >= 0 && tail == nil; i-- {
		if _, empty := body.Stmts[i].(*syntax.Nop); empty {
			continue
		}
		st, ok := body.Stmts[i].(*syntax.ExprStmt)
		if !ok {
			return false
		}
		tail = st
	}
	if tail == nil {
		return false
	}
	switch call := syntax.UnwrapParens(tail.Expr).(type) {
	case *syntax.FuncCall:
		fn := ctx.Types().ResolveFunction(call)
		if fn == nil || never(fn.Return, fn.DocReturn, fn.Inferred) {
			return true
		}
		decl := semanticquery.FunctionDecl(ctx.File, fn)
		return decl != nil && endless(decl.Body)
	case *syntax.MethodCall:
		if id, ok := call.Name.(*syntax.Identifier); ok {
			for _, class := range ctx.TypeOf(call.Var).Classes() {
				if m := ctx.Index().FindMethod(strings.TrimPrefix(class, `\`), id.Value, ctx.PHP); m != nil {
					return unknownMethod(ctx, m)
				}
			}
		}
		return true
	case *syntax.StaticCall:
		if id, ok := call.Name.(*syntax.Identifier); ok {
			if m := ctx.Index().FindMethod(semanticquery.StaticCallClass(ctx, call.Class), id.Value, ctx.PHP); m != nil {
				return unknownMethod(ctx, m)
			}
		}
		return true
	}
	return false
}

// unknownMethod reports whether a resolved method has no body to summarise
// (abstract or interface method) or never returns to its caller.
func unknownMethod(ctx *analysis.Context, m *index.Method) bool {
	if m.Abstract || never(m.Return, m.DocReturn, m.Inferred) {
		return true
	}
	decl := semanticquery.MethodDecl(ctx.File, ctx.Index(), m, ctx.PHP)
	return decl != nil && endless(decl.Body)
}

// endless reports whether a declaration body of this file never completes
// normally: it ends abruptly on every path and holds no return of its own.
func endless(body *syntax.Block) bool {
	if body == nil || !syntax.Terminates(body) {
		return false
	}
	returns := false
	syntax.Inspect(body, func(n syntax.Node) bool {
		switch n.(type) {
		case *syntax.Function, *syntax.Closure, *syntax.ArrowFunction, *syntax.ClassLike:
			return false
		case *syntax.Return:
			returns = true
		}
		return !returns
	})
	return !returns
}

// never reports whether one of the declared, documented or body-derived
// return types is `never`.
func never(returns ...string) bool {
	for _, t := range returns {
		if strings.EqualFold(strings.TrimPrefix(t, `\`), "never") {
			return true
		}
	}
	return false
}
