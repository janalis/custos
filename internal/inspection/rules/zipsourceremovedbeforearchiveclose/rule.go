// Package zipsourceremovedbeforearchiveclose implements the native ZipSourceRemovedBeforeArchiveClose inspection.
package zipsourceremovedbeforearchiveclose

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Keep ZIP source files until the archive closes."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ZipSourceRemovedBeforeArchiveClose" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "unlink") {
		return
	}
	path, ok := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "filename"))
	if !ok {
		return
	}
	// A removal inside a callback runs when the callback is invoked, not where it is written.
	for p, scope := c.Parent(), syntax.EnclosingVariableScope(c); p != scope; p = p.Parent() {
		branch, ok := p.(*syntax.If)
		if !ok || !branch.Body.Span().Contains(c.Span()) {
			continue
		}
		add, ok := syntax.UnwrapParens(branch.Cond).(*syntax.MethodCall)
		if !ok || !semanticquery.NativeMethod(ctx, add, "ZipArchive", "addFile") {
			continue
		}
		source, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(add.Args, 0, "filepath"))
		if !known || source != path {
			continue
		}
		flags := semanticquery.CallArgument(add.Args, 4, "flags")
		if flags != nil {
			continue
		}
		if !closedBefore(ctx, c, add.Var) {
			ctx.ReportNode(c, message)
		}
		return
	}
}

// closedBefore reports whether a close on the same archive object is evaluated
// before at on every ordinary path, including a close used as a condition.
func closedBefore(ctx *analysis.Context, at syntax.Node, receiver syntax.Expr) bool {
	target := ctx.Flow().Value(receiver)
	// A property or factory result has no flow identity; its spelling is compared instead.
	spelled := target.Identity == 0
	known := target.Complete && target.Invalidated != ^uint32(0)
	for _, call := range ctx.Flow().Calls(syntax.EnclosingVariableScope(at)) {
		if call.Receiver == nil || !strings.EqualFold(call.Name, "close") || call.Node.Span().End > at.Span().Start || !runsBefore(call.Node, at) {
			continue
		}
		value := ctx.Flow().Value(call.Receiver)
		if value.Identity != target.Identity {
			continue
		}
		if spelled && semanticquery.ExpansionCSame(ctx, call.Receiver, receiver) && !rebound(ctx, receiver, call.Node) {
			return true
		}
		if !spelled && known && value.Complete && value.Invalidated == target.Invalidated {
			return true
		}
	}
	return false
}

// rebound reports whether the receiver spelling may name another object at
// the later call: it is not a variable or property chain, or the receiver or
// an object it is read from is assigned between the two.
func rebound(ctx *analysis.Context, receiver syntax.Expr, later syntax.Node) bool {
	chain := []syntax.Expr{}
	for e := receiver; e != nil; {
		chain = append(chain, e)
		switch x := e.(type) {
		case *syntax.PropertyFetch:
			e = x.Var
		case *syntax.Variable, *syntax.StaticPropertyFetch:
			e = nil
		default:
			return true
		}
	}
	written := false
	syntax.InspectFile(ctx.File, func(n syntax.Node) bool {
		assign, ok := n.(*syntax.Assign)
		if ok && assign.Span().Start > receiver.Span().End && assign.Span().End < later.Span().Start {
			for _, e := range chain {
				written = written || semanticquery.ExpansionCSame(ctx, assign.Var, e)
			}
		}
		return !written
	})
	return written
}

// runsBefore proves that the textually earlier node is evaluated whenever at
// is: every construct between it and their common ancestor must run it
// unconditionally.
func runsBefore(earlier, at syntax.Node) bool {
	for node := earlier; node.Parent() != nil; node = node.Parent() {
		shared := node.Parent().Span().Contains(at.Span())
		switch parent := node.Parent().(type) {
		case *syntax.Ternary:
			if node != parent.Cond {
				return false
			}
		case *syntax.Binary:
			if node == parent.Right && shortCircuit(parent.Op.Kind) {
				return false
			}
		case *syntax.Assign:
			if node == parent.Value && parent.Op.Kind == syntax.TCoalesceEqual {
				return false
			}
		case *syntax.If:
			if node != parent.Cond {
				return false
			}
		case *syntax.ElseIf:
			if node != parent.Cond {
				return false
			}
		case *syntax.Try:
			if node != syntax.Node(parent.Body) {
				break
			}
			inFinally := parent.Finally != nil && parent.Finally.Span().Contains(at.Span())
			// A catch body runs when the protected block stopped early.
			if shared && !inFinally {
				return false
			}
			// A finally block, and the code after a try with catch blocks, also
			// runs when an earlier protected statement threw before the close.
			if (inFinally || len(parent.Catches) > 0) && !parent.Body.Stmts[0].Span().Contains(earlier.Span()) {
				return false
			}
		case *syntax.ExprStmt, *syntax.Block, *syntax.Finally:
		case syntax.Stmt, *syntax.MatchArm, *syntax.Closure, *syntax.ArrowFunction:
			return false
		}
		if shared {
			return true
		}
	}
	return true
}

func shortCircuit(kind syntax.TokenKind) bool {
	switch kind {
	case syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr, syntax.TCoalesce:
		return true
	}
	return false
}
