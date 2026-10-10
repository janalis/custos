// Package semaphorefailureenterscriticalsection implements the native SemaphoreFailureEntersCriticalSection inspection.
package semaphorefailureenterscriticalsection

import (
	"custos/internal/inspection/analysis"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Check semaphore acquisition before shared-state mutation."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SemaphoreFailureEntersCriticalSection" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "shm_put_var") && !semanticquery.NativeBuiltin(ctx, c, "shm_remove_var") {
		return
	}
	calls := ctx.Flow().Calls(syntax.EnclosingVariableScope(n))
	if len(calls) > 256 {
		return
	}
	for _, record := range calls {
		acquire, ok := record.Node.(*syntax.FuncCall)
		if !ok || !semanticquery.NativeBuiltin(ctx, acquire, "sem_acquire") || !semanticquery.NativeDominates(acquire, c) {
			continue
		}
		if _, discarded := acquire.Parent().(*syntax.ExprStmt); !discarded {
			continue
		}
		nonblocking, known := semanticquery.NativeTruth(ctx, semanticquery.CallArgument(acquire.Args, 1, "non_blocking"))
		if !known || !nonblocking {
			continue
		}
		sem := semanticquery.CallArgument(acquire.Args, 0, "semaphore")
		if !semanticquery.ExpansionDUnaliased(ctx, sem, n) {
			continue
		}
		for _, later := range calls {
			release, ok := later.Node.(*syntax.FuncCall)
			if !ok || !semanticquery.NativeBuiltin(ctx, release, "sem_release") || !semanticquery.NativeDominates(c, release) {
				continue
			}
			if semanticquery.NativeSameValue(ctx, sem, semanticquery.CallArgument(release.Args, 0, "semaphore")) {
				ctx.ReportNode(n, message)
				return
			}
		}
	}
}
