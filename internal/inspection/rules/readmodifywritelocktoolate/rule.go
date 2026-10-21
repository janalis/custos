// Package readmodifywritelocktoolate implements the native ReadModifyWriteLockTooLate inspection.
package readmodifywritelocktoolate

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Hold the lock across the complete read-modify-write operation."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ReadModifyWriteLockTooLate" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)

	if name != "file_put_contents" {
		return
	}
	flags := semanticquery.CallArgument(c.Args, 2, "flags")
	present, known := semanticquery.NativeFlagContains(ctx, flags, "LOCK_EX")
	if !known || !present {
		return
	}
	data := semanticquery.CallArgument(c.Args, 1, "data")
	if data == nil {
		return
	}
	path := semanticquery.CallArgument(c.Args, 0, "filename")
	// The earliest matching read is the one a spanning lock has to precede.
	var first *syntax.FuncCall
	syntax.Inspect(data, func(child syntax.Node) bool {
		expr, ok := child.(syntax.Expr)
		if !ok {
			return true
		}
		value := semanticquery.NativeValue(ctx, expr)
		for {
			unary, ok := value.(*syntax.Unary)
			if !ok || !unary.Op.Kind.IsCast() {
				break
			}
			value = unary.Expr
		}
		if read, ok := value.(*syntax.FuncCall); ok && semanticquery.NativeBuiltinName(ctx, read) == "file_get_contents" && semanticquery.SameOperand(ctx, path, semanticquery.CallArgument(read.Args, 0, "filename")) {
			if first == nil || read.Span().Start < first.Span().Start {
				first = read
			}
		}
		return true
	})

	if first != nil && !readModifyWriteLockSpans(ctx, c, first, path) {
		ctx.ReportNode(c, message)
	}
}

// lockFollows reports whether write can only be reached by completing
// statement: it comes later in the statement list that owns statement. Case
// lists are refused because another label can enter below the statement.
func lockFollows(statement syntax.Stmt, write syntax.Node) bool {
	if write.Span().Start < statement.Span().End {
		return false
	}
	owner := statement.Parent()
	_, list := syntax.StmtListOf(owner)
	_, label := owner.(*syntax.Case)
	return list && !label && lockWithin(write, owner)
}

// readModifyWriteLockDelegated reports an application call completed before
// read that may serialise the operation: its name mentions a lock without
// being an unlock, or it receives the path that is read and written.
func readModifyWriteLockDelegated(ctx *analysis.Context, callee syntax.Expr, args *syntax.ArgList, read *syntax.FuncCall, path syntax.Expr) bool {
	if callee == nil || args == nil || args.Span().End > read.Span().Start {
		return false
	}
	name := strings.ToLower(string(ctx.File.Src[callee.Span().Start:callee.Span().End]))
	if strings.Contains(name, "lock") && !strings.Contains(name, "unlock") {
		return true
	}
	for _, node := range args.Args {
		if arg, ok := node.(*syntax.Arg); ok && semanticquery.SameOperand(ctx, path, arg.Value) {
			return true
		}
	}
	return false
}

// readModifyWriteLockReleased reports whether release (an unlock, a downgrade
// to a shared lock or a close) ends lock on every path reaching write: it is
// a completed statement before the write, or it sits in the statement list
// that also took the lock. A release in a branch that leaves is ignored.
func readModifyWriteLockReleased(release, lock *syntax.FuncCall, write syntax.Node) bool {
	// Every expression of a scope is owned by a statement.
	n := release.Parent()
	statement, ok := n.(syntax.Stmt)
	for !ok {
		n = n.Parent()
		statement, ok = n.(syntax.Stmt)
	}
	owner := statement.Parent()
	_, list := syntax.StmtListOf(owner)
	return lockFollows(statement, write) || (list && lockWithin(lock, owner))
}

// readModifyWriteLockSpans reports whether write may run under a lock that
// was already held when read ran: an exclusive flock (or one whose operation
// is unknown) taken before the read and not provably released before the
// write, or an earlier application call that may take the lock itself.
func readModifyWriteLockSpans(ctx *analysis.Context, write, read *syntax.FuncCall, path syntax.Expr) bool {
	scope := syntax.EnclosingVariableScope(write)
	held := map[string]*syntax.FuncCall{}
	delegated := false
	visit := func(n syntax.Node) bool {
		if n.Span().Start >= write.Span().Start || (n != scope && syntax.IsVariableScope(n)) {
			return false
		}
		switch call := n.(type) {
		case *syntax.MethodCall:
			delegated = delegated || readModifyWriteLockDelegated(ctx, call.Name, call.Args, read, path)
		case *syntax.StaticCall:
			delegated = delegated || readModifyWriteLockDelegated(ctx, call.Name, call.Args, read, path)
		case *syntax.FuncCall:
			name := semanticquery.NativeBuiltinName(ctx, call)
			handle := semanticquery.CallArgument(call.Args, 0, "stream")
			switch {
			case name == "":
				delegated = delegated || readModifyWriteLockDelegated(ctx, call.Name, call.Args, read, path)
			case handle == nil || (name != "flock" && name != "fclose"):
			default:
				key := string(ctx.File.Src[handle.Span().Start:handle.Span().End])
				// LOCK_UN (3) shares a bit with LOCK_EX (2), so compare the
				// whole operation once the optional LOCK_NB (4) is masked out.
				operation, known := semanticquery.FileLockOperation(ctx, semanticquery.CallArgument(call.Args, 1, "operation"))
				if name == "flock" && (!known || operation&^4 == 2) {
					if call.Span().End <= read.Span().Start {
						held[key] = call
					}
				} else if lock := held[key]; lock != nil && readModifyWriteLockReleased(call, lock, write) {
					delete(held, key)
				}
			}
		}
		return true
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return delegated || len(held) > 0
}

func lockWithin(n, ancestor syntax.Node) bool {
	for ; n != nil; n = n.Parent() {
		if n == ancestor {
			return true
		}
	}
	return false
}
