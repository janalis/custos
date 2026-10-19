package semanticquery

import (
	"cmp"
	"slices"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// PriorCall records a possible operation before a use. Must distinguishes an
// unconditional predecessor from an operation reached on only some paths.
type PriorCall struct {
	Node syntax.Node
	Must bool
}

// CallHistory retains conditional operations instead of discarding them.
// Complete establishes the receiver identity and the bounded history of the
// requested operation names; it does not certify effects of other operations.
// An incomplete history may still retain proven operations in the current epoch.
type CallHistory struct {
	Events   []PriorCall
	Complete bool
}

const maxHistoryWork = 4096

// Certain returns a replayable list only when every recorded operation must
// precede the use. An empty known history differs from an unknown history.
func (h CallHistory) Certain() ([]syntax.Node, bool) {
	if !h.Complete {
		return nil, false
	}
	for _, event := range h.Events {
		if !event.Must {
			return nil, false
		}
	}
	nodes := make([]syntax.Node, 0, len(h.Events))
	for _, event := range h.Events {
		nodes = append(nodes, event.Node)
	}
	return nodes, true
}

// NativeCallHistory orders matching receiver operations by evaluation completion.
// An operation in another call's arguments executes before its containing call.
// Possible operations on earlier loop turns, unknown receiver aliases and
// exhausted work bounds prevent an unconditional replay proof.
func NativeCallHistory(ctx *analysis.Context, at syntax.Node, receiver syntax.Expr, names ...string) CallHistory {
	return nativeCallHistory(ctx, at, receiver, maxHistoryWork, names...)
}

func nativeCallHistory(ctx *analysis.Context, at syntax.Node, receiver syntax.Expr, budget int, names ...string) (history CallHistory) {
	history = CallHistory{Complete: true}
	defer func() {
		slices.SortStableFunc(history.Events, func(a, b PriorCall) int {
			return cmp.Compare(a.Node.Span().End, b.Node.Span().End)
		})
	}()
	// Identity agrees on every path independently of taint-source completeness.
	// A fresh handle may retain incomplete input data without losing its identity.
	target := ctx.Flow().Value(receiver)
	if target.Identity == 0 || target.Invalidated == ^uint32(0) || !nativeHandleUnescaped(ctx, receiver, at, true) {
		return CallHistory{}
	}
	history.Complete = target.Invalidated == 0
	scope := syntax.EnclosingVariableScope(at)
	work := 0
	for _, call := range ctx.Flow().Calls(scope) {
		work++
		if work > budget {
			ctx.Gap("query-budget", at.Span())
			history.Complete = false
			break
		}
		matches := false
		for _, name := range names {
			matches = matches || strings.EqualFold(call.Name, name)
		}
		if !matches || call.Node == at {
			continue
		}
		// An unpacked list that may name the receiver lets it escape, which
		// ends the history before this loop.
		candidate := call.Receiver
		if candidate == nil {
			if function, ok := call.Node.(*syntax.FuncCall); ok {
				candidate = CallArgument(function.Args, 0, "")
			}
		}
		if candidate == nil {
			continue
		}
		value := ctx.Flow().Value(candidate)
		if value.Identity != 0 && value.Identity != target.Identity {
			continue
		}
		before := call.Node.Span().End <= at.Span().Start
		if !before {
			for parent := call.Node.Parent(); parent != nil && !syntax.IsVariableScope(parent); parent = parent.Parent() {
				work++
				if work > budget {
					ctx.Gap("query-budget", at.Span())
					history.Complete = false
					return history
				}
				switch parent.(type) {
				case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
					if parent.Span().Contains(at.Span()) && !historyAllocationRestarts(receiver, target.Expr, target.Identity, parent, at) {
						history.Complete = false
						return history
					}
				}
			}
			continue
		}
		if value.Identity == 0 || value.Invalidated != target.Invalidated {
			history.Complete = false
			continue
		}
		history.Events = append(history.Events, PriorCall{Node: call.Node, Must: NativeDominates(call.Node, at)})
	}

	return history
}

// historyAllocationRestarts proves a plain local receives a fresh allocation
// in this loop body before each use. Loop initializers and conditional writes
// do not reset an object's history on every turn.
func historyAllocationRestarts(receiver, origin syntax.Expr, identity uint32, loop, at syntax.Node) bool {
	variable, plain := syntax.UnwrapParens(receiver).(*syntax.Variable)
	origin = syntax.UnwrapParens(origin)
	if !plain || variable.NameExpr != nil || origin == nil || origin.Span().Start+1 != identity {
		return false
	}
	child := syntax.Node(origin)
	for {
		parent, ok := child.Parent().(*syntax.Paren)
		if !ok {
			break
		}
		child = parent
	}
	assignment, assigned := child.Parent().(*syntax.Assign)
	if !assigned || assignment.Op.Kind != syntax.TEqual || assignment.ByRef || assignment.Value != child {
		return false
	}
	target, local := assignment.Var.(*syntax.Variable)
	if !local || target.NameExpr != nil || target.Name != variable.Name {
		return false
	}
	var body syntax.Stmt
	switch statement := loop.(type) {
	case *syntax.For:
		body = statement.Body
	case *syntax.Foreach:
		body = statement.Body
	case *syntax.While:
		body = statement.Body
	case *syntax.DoWhile:
		body = statement.Body
	}
	return body != nil && body.Span().Contains(assignment.Span()) && NativeDominates(assignment, at)
}
