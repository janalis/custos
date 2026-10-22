// Package filetruncatedbeforelock implements the native FileTruncatedBeforeLock inspection.
package filetruncatedbeforelock

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Acquire the lock before truncating the file."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FileTruncatedBeforeLock" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)
	if name != "fopen" {
		return
	}
	mode, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 1, "mode"))
	if !known || !strings.HasPrefix(mode, "w") {
		return
	}
	assignment, ok := c.Parent().(*syntax.Assign)
	if !ok {
		return
	}
	variable, ok := assignment.Var.(*syntax.Variable)
	if !ok {
		return
	}
	for _, later := range ctx.Flow().Calls(syntax.EnclosingVariableScope(c)) {
		lock, ok := later.Node.(*syntax.FuncCall)
		if !ok || later.Name != "flock" || lock.Span().Start <= c.Span().End {
			continue
		}
		arg, ok := semanticquery.CallArgument(lock.Args, 0, "stream").(*syntax.Variable)
		if !ok || arg.Name != variable.Name || ctx.Flow().Value(arg).Identity != ctx.Flow().Value(c).Identity {
			continue
		}
		// LOCK_UN (3) shares a bit with LOCK_EX (2), so compare the whole
		// operation once the optional LOCK_NB (4) modifier is masked out.
		if operation, known := semanticquery.FileLockOperation(ctx, semanticquery.CallArgument(lock.Args, 1, "operation")); known && operation&^4 == 2 {
			ctx.ReportNode(c, message)
			return
		}
	}
}
