// Package filelockfailureunchecked implements the native FileLockFailureUnchecked inspection.
package filelockfailureunchecked

import (
	"sort"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/flow"
)

const message = "Acquire the lock successfully before writing."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FileLockFailureUnchecked" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)
	if name != "fwrite" && name != "ftruncate" {
		return
	}
	handle := semanticquery.CallArgument(c.Args, 0, "stream")
	if handle == nil {
		return
	}
	state, known := ctx.Flow().StateBefore(c, handle, "flock")
	if !known {
		state, known = assertionLockAttempt(ctx, c, handle)
	}
	if !known || state.Successful {
		return
	}
	// The flow state only records a direct truthiness test of the call; accept
	// every other test of its result that the write provably depends on.
	scope := syntax.EnclosingVariableScope(c)
	for _, reached := range ctx.Flow().Calls(scope) {
		lock, ok := reached.Node.(*syntax.FuncCall)
		if ok && lock.Span() == state.Span && semanticquery.NativeCallSucceeded(ctx, lock, c) {
			return
		}
	}
	ctx.ReportNode(c, message)
}

// assertionLockAttempt retains a may-attempt witness when disabled assertions
// join the executed lock call with the skipped path. It supplies no proof of
// acquisition and does not weaken StateBefore's common-operation contract.
func assertionLockAttempt(ctx *analysis.Context, write *syntax.FuncCall, handle syntax.Expr) (flow.State, bool) {
	current := ctx.Flow().Value(handle)
	if !current.Complete || current.Identity == 0 || current.Invalidated == ^uint32(0) {
		return flow.State{}, false
	}
	scopes := ctx.Memo("file-lock.assertion-attempts", func() any { return map[syntax.Node]*assertionLockIndex{} }).(map[syntax.Node]*assertionLockIndex)
	scope := syntax.EnclosingVariableScope(write)
	ix := scopes[scope]
	if ix == nil {
		ix = &assertionLockIndex{attempts: map[uint32][]assertionLockCandidate{}, ordered: append([]flow.Call(nil), ctx.Flow().Calls(scope)...)}
		sort.Slice(ix.ordered, func(i, j int) bool { return ix.ordered[i].Node.Span().End < ix.ordered[j].Node.Span().End })
		for _, reached := range ix.ordered {
			lock, ok := reached.Node.(*syntax.FuncCall)
			if !ok || reached.Name != "flock" || !semanticquery.NativeBuiltin(ctx, lock, "flock") {
				continue
			}
			nested := false
			for p := lock.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
				if outer, ok := p.(*syntax.FuncCall); ok && ctx.Types().BuiltinAssert(outer) {
					nested = true
					break
				}
			}
			if !nested {
				continue
			}
			operation, known := semanticquery.FileLockOperation(ctx, semanticquery.CallArgument(lock.Args, 1, "operation"))
			if !known || operation&^4 != 2 {
				continue
			}
			stream := semanticquery.CallArgument(lock.Args, 0, "stream")
			if stream == nil {
				continue
			}
			prior := ctx.Flow().Value(stream)
			if prior.Complete && prior.Identity != 0 && prior.Invalidated != ^uint32(0) {
				ix.attempts[prior.Identity] = append(ix.attempts[prior.Identity], assertionLockCandidate{lock, prior.Invalidated})
			}
		}
		scopes[scope] = ix
	}
	for _, candidate := range ix.attempts[current.Identity] {
		ix.work++
		if ix.work > flow.MaxTransfers {
			return flow.State{}, false
		}
		lock := candidate.call
		if candidate.invalidated != current.Invalidated || !semanticquery.NativeDominates(lock, write) {
			continue
		}
		start := sort.Search(len(ix.ordered), func(i int) bool { return ix.ordered[i].Node.Span().End > lock.Span().End })
		interrupted := false
		for _, between := range ix.ordered[start:] {
			if between.Node.Span().End > write.Span().Start {
				break
			}
			ix.work++
			if ix.work > flow.MaxTransfers {
				return flow.State{}, false
			}
			if outer, ok := between.Node.(*syntax.FuncCall); ok && ctx.Types().BuiltinAssert(outer) {
				continue
			}
			interrupted = true
			break
		}
		if !interrupted {
			return flow.State{Operation: "flock", Span: lock.Span()}, true
		}
	}
	return flow.State{}, false
}

type assertionLockCandidate struct {
	call        *syntax.FuncCall
	invalidated uint32
}

type assertionLockIndex struct {
	attempts map[uint32][]assertionLockCandidate
	ordered  []flow.Call
	work     int
}
