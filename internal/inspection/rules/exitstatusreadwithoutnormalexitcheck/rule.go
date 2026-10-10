// Package exitstatusreadwithoutnormalexitcheck implements the native ExitStatusReadWithoutNormalExitCheck inspection.
package exitstatusreadwithoutnormalexitcheck

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Check normal termination before decoding the exit code."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ExitStatusReadWithoutNormalExitCheck" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "pcntl_wexitstatus") {
		return
	}
	input := semanticquery.CallArgument(c.Args, 0, "status")
	wait := semanticquery.ExpansionCOutputCall(ctx, c, input, []string{"pcntl_waitpid"}, 1, "status")
	if wait == nil {
		wait = semanticquery.ExpansionCOutputCall(ctx, c, input, []string{"pcntl_wait"}, 0, "status")
	}
	if wait == nil {
		return
	}
	for parent := c.Parent(); parent != nil; parent = parent.Parent() {
		branch, ok := parent.(*syntax.If)
		if !ok || !branch.Body.Span().Contains(c.Span()) {
			continue
		}
		exited, success := false, false
		var inspect func(syntax.Expr)
		inspect = func(e syntax.Expr) {
			e = syntax.UnwrapParens(e)
			if b, ok := e.(*syntax.Binary); ok {
				if b.Op.Kind == syntax.TBooleanAnd {
					inspect(b.Left)
					inspect(b.Right)
					return
				}
				value, known := semanticquery.NativeInt(ctx, b.Right)
				origin := semanticquery.NativeValue(ctx, b.Left)
				if known && origin == wait && ((b.Op.Kind == syntax.TGreater && value == 0) || (b.Op.Kind == syntax.TIsNotIdentical && value == -1 && semanticquery.CallArgument(wait.Args, 2, "flags") == nil)) {
					success = true
				}
			}
			if guard, ok := e.(*syntax.FuncCall); ok && semanticquery.NativeBuiltin(ctx, guard, "pcntl_wifexited") && semanticquery.ExpansionCSame(ctx, input, semanticquery.CallArgument(guard.Args, 0, "status")) {
				exited = true
			}
		}
		inspect(branch.Cond)
		if success && exited {
			return
		}
	}

	ctx.ReportNode(c, message)
}
