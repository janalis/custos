package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

func nativeSameExpression(ctx *analysis.Context, a, b syntax.Expr) bool {
	if a == nil || b == nil {
		return false
	}
	av, bv := ctx.Flow().Value(a), ctx.Flow().Value(b)
	if !av.Complete || !bv.Complete {
		return false
	}
	if av.Expr != nil && av.Expr == bv.Expr {
		return true
	}
	if left, ok := NativeString(ctx, a); ok {
		right, known := NativeString(ctx, b)
		return known && left == right
	}
	return av.Identity != 0 && av.Identity == bv.Identity && av.Invalidated == bv.Invalidated && av.Invalidated != ^uint32(0)
}

func nativePriorGlobal(ctx *analysis.Context, at syntax.Node, names ...string) []*syntax.FuncCall {
	out := []*syntax.FuncCall{}
	for _, c := range ctx.Flow().Calls(syntax.EnclosingVariableScope(at)) {
		call, ok := c.Node.(*syntax.FuncCall)
		if !ok || !NativeDominates(call, at) {
			continue
		}
		for _, name := range names {
			if c.Name == name {
				out = append(out, call)
				break
			}
		}
	}
	return out
}

// nativeDirectoryGuard proves that the canonical parent of the temporary file
// held by path was compared with a canonical directory after source created it
// and before at runs: either an earlier terminating mismatch branch, or an
// enclosing branch entered only on a match.
func nativeDirectoryGuard(ctx *analysis.Context, at syntax.Node, path syntax.Expr, source syntax.Node) bool {
	variable, ok := syntax.UnwrapParens(path).(*syntax.Variable)
	if !ok {
		return false
	}
	for child, p := at, at.Parent(); p != nil && !syntax.IsVariableScope(p); child, p = p, p.Parent() {
		switch parent := p.(type) {
		case *syntax.If:
			if child == syntax.Node(parent.Body) && parent.Cond.Span().Start >= source.Span().End && nativeDirectoryComparison(ctx, parent.Cond, variable.Name, true) {
				return true
			}
		case *syntax.Block:
			for _, st := range parent.Stmts {
				branch, ok := st.(*syntax.If)
				if !ok || branch.Span().End > at.Span().Start || branch.Span().Start < source.Span().End || !syntax.Terminates(branch.Body) {
					continue
				}
				if nativeDirectoryComparison(ctx, branch.Cond, variable.Name, false) {
					return true
				}
			}
		}
	}
	return false
}

// nativeDirectoryComparison matches realpath(dirname($name)) compared with
// another realpath() call in either operand order. With same it accepts the
// equality forms, also as a conjunct; otherwise the inequality forms, also as
// a disjunct. Canonical paths are absolute strings or false, so the loose
// operators decide exactly like the strict ones.
func nativeDirectoryComparison(ctx *analysis.Context, cond syntax.Expr, name string, same bool) bool {
	binary, ok := syntax.UnwrapParens(cond).(*syntax.Binary)
	if !ok {
		return false
	}
	join, strict, loose := syntax.TBooleanOr, syntax.TIsNotIdentical, syntax.TIsNotEqual
	if same {
		join, strict, loose = syntax.TBooleanAnd, syntax.TIsIdentical, syntax.TIsEqual
	}
	switch binary.Op.Kind {
	case join:
		return nativeDirectoryComparison(ctx, binary.Left, name, same) || nativeDirectoryComparison(ctx, binary.Right, name, same)
	case strict, loose:
		return nativeCanonicalParent(ctx, binary.Left, name) && nativeCanonicalPath(ctx, binary.Right) != nil ||
			nativeCanonicalParent(ctx, binary.Right, name) && nativeCanonicalPath(ctx, binary.Left) != nil
	}
	return false
}

// nativeCanonicalPath returns the argument of a builtin realpath() call.
func nativeCanonicalPath(ctx *analysis.Context, e syntax.Expr) syntax.Expr {
	call, ok := syntax.UnwrapParens(e).(*syntax.FuncCall)
	if !ok || NativeBuiltinName(ctx, call) != "realpath" {
		return nil
	}
	return CallArgument(call.Args, 0, "path")
}

// nativeCanonicalParent matches realpath(dirname($name)).
func nativeCanonicalParent(ctx *analysis.Context, e syntax.Expr, name string) bool {
	dir, ok := syntax.UnwrapParens(nativeCanonicalPath(ctx, e)).(*syntax.FuncCall)
	if !ok || NativeBuiltinName(ctx, dir) != "dirname" {
		return false
	}
	v, ok := syntax.UnwrapParens(CallArgument(dir.Args, 0, "path")).(*syntax.Variable)
	return ok && v.Name == name
}

// CheckDirectoryIteratorDotEntries checks the independently specified contract.
func CheckDirectoryIteratorDotEntries(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.FuncCall)
	if !ok {
		return
	}
	name := NativeBuiltinName(ctx, c)
	report := false

	if name != "unlink" && name != "rmdir" {
		return
	}
	path, ok := CallArgument(c.Args, 0, "filename").(*syntax.MethodCall)
	if !ok {
		return
	}
	variable, ok := path.Var.(*syntax.Variable)
	if !ok {
		return
	}
	for p := c.Parent(); p != nil; p = p.Parent() {
		loop, ok := p.(*syntax.Foreach)
		if !ok {
			continue
		}
		iterator, ok := loop.Expr.(*syntax.New)
		if !ok {
			continue
		}
		class, ok := iterator.Class.(*syntax.Name)
		if !ok || ctx.Names().Class(class.Value, class.Span().Start) != "DirectoryIterator" {
			continue
		}
		value, ok := loop.Value.(*syntax.Variable)
		if !ok || value.Name != variable.Name {
			continue
		}
		if nativeIteratorReassigned(loop, c, variable.Name) {
			return
		}
		report = nativeDotEntriesExcludedAt(ctx, loop, c, variable.Name) != nativeDotEntryNames
		break
	}

	if report {
		ctx.ReportNode(c, message)
	}
}

// nativeDotEntryNames is the mask of both dot entries: bit 1 for ".", bit 2
// for "..".
const nativeDotEntryNames uint8 = 3

// nativeDotEntriesExcludedAt returns which dot entries the loop variable can
// no longer name when c runs: early exits completed before c contribute the
// false outcome of their condition, enclosing branches and short-circuit
// operands the outcome that lets control reach c.
func nativeDotEntriesExcludedAt(ctx *analysis.Context, loop *syntax.Foreach, c *syntax.FuncCall, variable string) uint8 {
	var mask uint8
	syntax.Inspect(loop.Body, func(child syntax.Node) bool {
		branch, ok := child.(*syntax.If)
		if ok && branch.Span().End <= c.Span().Start && syntax.Terminates(branch.Body) {
			mask |= nativeDotEntriesExcluded(ctx, branch.Cond, variable, false)
		}
		return true
	})
	for child, ancestor := syntax.Node(c), c.Parent(); ancestor != loop; child, ancestor = ancestor, ancestor.Parent() {
		switch a := ancestor.(type) {
		case *syntax.If:
			if child == syntax.Node(a.Body) {
				mask |= nativeDotEntriesExcluded(ctx, a.Cond, variable, true)
			} else if child != syntax.Node(a.Cond) {
				mask |= nativeDotEntriesExcluded(ctx, a.Cond, variable, false)
			}
		case *syntax.ElseIf:
			if child == syntax.Node(a.Body) {
				mask |= nativeDotEntriesExcluded(ctx, a.Cond, variable, true)
			}
		case *syntax.Binary:
			if child != syntax.Node(a.Right) {
				continue
			}
			switch a.Op.Kind {
			case syntax.TBooleanAnd, syntax.TAnd:
				mask |= nativeDotEntriesExcluded(ctx, a.Left, variable, true)
			case syntax.TBooleanOr, syntax.TOr:
				mask |= nativeDotEntriesExcluded(ctx, a.Left, variable, false)
			}
		}
	}
	return mask
}

// nativeDotEntriesExcluded returns which dot entries the loop variable cannot
// name once cond has evaluated to truth.
func nativeDotEntriesExcluded(ctx *analysis.Context, cond syntax.Expr, variable string, truth bool) uint8 {
	switch e := cond.(type) {
	case *syntax.Paren:
		return nativeDotEntriesExcluded(ctx, e.Expr, variable, truth)
	case *syntax.Unary:
		if e.Op.Kind == syntax.TExclaim {
			return nativeDotEntriesExcluded(ctx, e.Expr, variable, !truth)
		}
	case *syntax.Binary:
		switch e.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd, syntax.TBooleanOr, syntax.TOr:
			left := nativeDotEntriesExcluded(ctx, e.Left, variable, truth)
			right := nativeDotEntriesExcluded(ctx, e.Right, variable, truth)
			// A true conjunction and a false disjunction fix both operands;
			// otherwise only what both operands agree on is known.
			if (e.Op.Kind == syntax.TBooleanAnd || e.Op.Kind == syntax.TAnd) == truth {
				return left | right
			}
			return left & right
		case syntax.TIsIdentical, syntax.TIsEqual:
			if !truth {
				return nativeDotNameCompared(ctx, e, variable)
			}
		case syntax.TIsNotIdentical, syntax.TIsNotEqual:
			if truth {
				return nativeDotNameCompared(ctx, e, variable)
			}
		}
	case *syntax.MethodCall:
		// "." and ".." are always directories, never regular files or links.
		switch nativeEntryMethod(e, variable) {
		case "isdot", "isdir":
			if !truth {
				return nativeDotEntryNames
			}
		case "isfile", "islink":
			if truth {
				return nativeDotEntryNames
			}
		}
	case *syntax.FuncCall:
		if truth || NativeBuiltinName(ctx, e) != "in_array" || !nativeEntryName(CallArgument(e.Args, 0, "needle"), variable) {
			return 0
		}
		var mask uint8
		if names := NativeArray(ctx, CallArgument(e.Args, 1, "haystack")); names != nil {
			for _, item := range names.Items {
				mask |= nativeDotNameBit(ctx, item.Value)
			}
		}
		return mask
	}
	return 0
}

// nativeEntryMethod returns the lower-cased name of a method called directly
// on the loop variable, or "".
func nativeEntryMethod(call *syntax.MethodCall, variable string) string {
	id, ok := call.Name.(*syntax.Identifier)
	receiver, rok := call.Var.(*syntax.Variable)
	if !ok || !rok || receiver.Name != variable {
		return ""
	}
	return strings.ToLower(id.Value)
}

// nativeEntryName reports whether e reads the entry's own name, which is "."
// or ".." exactly for the dot entries.
func nativeEntryName(e syntax.Expr, variable string) bool {
	call, ok := e.(*syntax.MethodCall)
	if !ok || (call.Args != nil && len(call.Args.Args) != 0) {
		return false
	}
	name := nativeEntryMethod(call, variable)
	return name == "getfilename" || name == "getbasename"
}

func nativeDotNameCompared(ctx *analysis.Context, e *syntax.Binary, variable string) uint8 {
	if nativeEntryName(e.Left, variable) {
		return nativeDotNameBit(ctx, e.Right)
	}
	if nativeEntryName(e.Right, variable) {
		return nativeDotNameBit(ctx, e.Left)
	}
	return 0
}

func nativeDotNameBit(ctx *analysis.Context, e syntax.Expr) uint8 {
	switch name, _ := NativeString(ctx, e); name {
	case ".":
		return 1
	case "..":
		return 2
	}
	return 0
}

// CheckFileLockFailureUnchecked checks the independently specified contract.
func CheckFileLockFailureUnchecked(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.FuncCall)
	if !ok {
		return
	}
	name := NativeBuiltinName(ctx, c)
	if name != "fwrite" && name != "ftruncate" {
		return
	}
	handle := CallArgument(c.Args, 0, "stream")
	if handle == nil {
		return
	}
	state, known := ctx.Flow().StateBefore(c, handle, "flock")
	if !known || state.Successful {
		return
	}
	// The flow state only records a direct truthiness test of the call; accept
	// every other test of its result that the write provably depends on.
	scope := syntax.EnclosingVariableScope(c)
	for _, reached := range ctx.Flow().Calls(scope) {
		lock, ok := reached.Node.(*syntax.FuncCall)
		if ok && lock.Span() == state.Span && nativeLockResultTested(ctx, lock, c, scope) {
			return
		}
	}
	ctx.ReportNode(c, message)
}

// nativeLockTruth records what the truthiness of an expression proves about a
// lock call: held when it is truthy, held when it is falsy, or both.
type nativeLockTruth struct{ whenTrue, whenFalse bool }

// nativeLockResultTested proves that write runs only after lock returned true,
// through a test of the call itself or of an unchanged local holding its result.
func nativeLockResultTested(ctx *analysis.Context, lock *syntax.FuncCall, write, scope syntax.Node) bool {
	stored, held, guarded := nativeLockTruthGuards(lock, nativeLockTruth{whenTrue: true}, write)
	if guarded || stored == nil {
		return guarded
	}
	name := stored.Var.(*syntax.Variable).Name
	value := ctx.Flow().Value(stored.Value).Expr
	visit := func(n syntax.Node) bool {
		if guarded || (n != scope && syntax.IsVariableScope(n)) {
			return false
		}
		// A read whose single reaching value is the stored result tests the
		// same acquisition; a reassigned or merged local proves nothing.
		if use, ok := n.(*syntax.Variable); ok && use.Name == name && value != nil && ctx.Flow().Value(use).Expr == value {
			_, _, guarded = nativeLockTruthGuards(use, held, write)
		}
		return true
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return guarded
}

// nativeLockTruthGuards follows x outwards through negation, error
// suppression, boolean comparisons, short-circuit operators and a plain local
// assignment, and reports whether the construct consuming it lets write run
// only when the lock is held. The first plain assignment met is returned with
// the truth it stores, so later reads of that local can be followed too.
func nativeLockTruthGuards(x syntax.Expr, held nativeLockTruth, write syntax.Node) (stored *syntax.Assign, storedHeld nativeLockTruth, guarded bool) {
	for held.whenTrue || held.whenFalse {
		switch parent := x.Parent().(type) {
		case *syntax.Paren:
			x = parent
		case *syntax.Unary:
			switch parent.Op.Kind {
			case syntax.TExclaim:
				held = nativeLockTruth{held.whenFalse, held.whenTrue}
			case syntax.TAt:
			default:
				return stored, storedHeld, false
			}
			x = parent
		case *syntax.Assign:
			target, ok := parent.Var.(*syntax.Variable)
			if !ok || target.NameExpr != nil || x != parent.Value || parent.ByRef || parent.Op.Kind != syntax.TEqual {
				return stored, storedHeld, false
			}
			if stored == nil {
				stored, storedHeld = parent, held
			}
			x = parent
		case *syntax.Binary:
			switch parent.Op.Kind {
			case syntax.TIsIdentical, syntax.TIsEqual, syntax.TIsNotIdentical, syntax.TIsNotEqual:
				other := parent.Right
				if x == parent.Right {
					other = parent.Left
				}
				value, ok := nativeBoolConstant(other)
				if !ok {
					return stored, storedHeld, false
				}
				// Every followed expression is boolean, so `=== false`
				// and `!== true` are plain negations.
				if value == (parent.Op.Kind == syntax.TIsNotIdentical || parent.Op.Kind == syntax.TIsNotEqual) {
					held = nativeLockTruth{held.whenFalse, held.whenTrue}
				}
			case syntax.TBooleanAnd, syntax.TAnd:
				if x == parent.Left && nativeLockWithin(write, parent.Right) {
					return stored, storedHeld, held.whenTrue
				}
				held.whenFalse = false
			case syntax.TBooleanOr, syntax.TOr:
				if x == parent.Left && nativeLockWithin(write, parent.Right) {
					return stored, storedHeld, held.whenFalse
				}
				// `flock(...) or die()` completes only with the lock held.
				_, throws := syntax.UnwrapParens(parent.Right).(*syntax.Throw)
				if x == parent.Left && held.whenTrue && (throws || syntax.ExitInvocation(parent.Right)) {
					held.whenFalse = true
				} else {
					held.whenTrue = false
				}
			default:
				return stored, storedHeld, false
			}
			x = parent
		case *syntax.Ternary:
			if x != parent.Cond {
				return stored, storedHeld, false
			}
			if parent.Then != nil && nativeLockWithin(write, parent.Then) {
				return stored, storedHeld, held.whenTrue
			}
			return stored, storedHeld, held.whenFalse && nativeLockWithin(write, parent.Else)
		case *syntax.ExprStmt:
			return stored, storedHeld, held.whenTrue && held.whenFalse && nativeLockFollows(parent, write)
		case *syntax.If:
			if nativeLockWithin(write, parent.Body) {
				return stored, storedHeld, held.whenTrue
			}
			if nativeLockWithin(write, parent) {
				return stored, storedHeld, held.whenFalse && write.Span().Start > parent.Body.Span().End
			}
			// Past the statement, the branch that means "not held" must have left.
			left := held.whenFalse && syntax.Terminates(parent.Body)
			if held.whenTrue && len(parent.ElseIfs) == 0 && parent.Else != nil {
				left = left || syntax.Terminates(parent.Else.Body)
			}
			return stored, storedHeld, left && nativeLockFollows(parent, write)
		case *syntax.ElseIf:
			return stored, storedHeld, held.whenTrue && nativeLockWithin(write, parent.Body)
		case *syntax.While:
			if nativeLockWithin(write, parent.Body) {
				return stored, storedHeld, held.whenTrue
			}
			return stored, storedHeld, held.whenFalse && !nativeLockLoopEscapes(parent.Body) && nativeLockFollows(parent, write)
		case *syntax.DoWhile:
			return stored, storedHeld, held.whenFalse && !nativeLockLoopEscapes(parent.Body) && nativeLockFollows(parent, write)
		default:
			return stored, storedHeld, false
		}
	}
	return stored, storedHeld, false
}

func nativeBoolConstant(e syntax.Expr) (value, ok bool) {
	constant, ok := syntax.UnwrapParens(e).(*syntax.ConstFetch)
	if !ok {
		return false, false
	}
	switch strings.ToLower(strings.TrimPrefix(constant.Name.Value, `\`)) {
	case "true":
		return true, true
	case "false":
		return false, true
	}
	return false, false
}

func nativeLockWithin(n, ancestor syntax.Node) bool {
	for ; n != nil; n = n.Parent() {
		if n == ancestor {
			return true
		}
	}
	return false
}

// nativeLockFollows reports whether write can only be reached by completing
// statement: it comes later in the statement list that owns statement. Case
// lists are refused because another label can enter below the statement.
func nativeLockFollows(statement syntax.Stmt, write syntax.Node) bool {
	if write.Span().Start < statement.Span().End {
		return false
	}
	owner := statement.Parent()
	_, list := syntax.StmtListOf(owner)
	_, label := owner.(*syntax.Case)
	return list && !label && nativeLockWithin(write, owner)
}

// nativeLockLoopEscapes reports a break that could leave a retry loop before
// its condition established the lock. Scopes using goto have no flow state.
func nativeLockLoopEscapes(body syntax.Stmt) bool {
	escapes := false
	syntax.Inspect(body, func(n syntax.Node) bool {
		_, leaves := n.(*syntax.Break)
		escapes = escapes || leaves
		return !escapes
	})
	return escapes
}

// CheckFileTruncatedBeforeLock checks the independently specified contract.
func CheckFileTruncatedBeforeLock(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.FuncCall)
	if !ok {
		return
	}
	name := NativeBuiltinName(ctx, c)
	report := false

	if name != "fopen" {
		return
	}
	mode, known := NativeString(ctx, CallArgument(c.Args, 1, "mode"))
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
		arg, ok := CallArgument(lock.Args, 0, "stream").(*syntax.Variable)
		if !ok || arg.Name != variable.Name || ctx.Flow().Value(arg).Identity != ctx.Flow().Value(c).Identity {
			continue
		}
		// LOCK_UN (3) shares a bit with LOCK_EX (2), so compare the whole
		// operation once the optional LOCK_NB (4) modifier is masked out.
		if operation, known := fileLockOperation(ctx, CallArgument(lock.Args, 1, "operation")); known && operation&^4 == 2 {
			report = true
		}
	}

	if report {
		ctx.ReportNode(c, message)
	}
}

// fileLockOperation folds a flock operation made of known integers and
// constants joined by bitwise or. Anything else stays unknown.
func fileLockOperation(ctx *analysis.Context, e syntax.Expr) (int64, bool) {
	if e == nil {
		return 0, false
	}
	if n, ok := NativeInt(ctx, e); ok {
		return n, true
	}
	if binary, ok := NativeValue(ctx, e).(*syntax.Binary); ok && binary.Op.Kind == syntax.TBar {
		left, lk := fileLockOperation(ctx, binary.Left)
		right, rk := fileLockOperation(ctx, binary.Right)
		return left | right, lk && rk
	}
	return 0, false
}

// CheckGlobFailureUnchecked checks the independently specified contract.
func CheckGlobFailureUnchecked(ctx *analysis.Context, n syntax.Node, message string) {
	switch x := n.(type) {
	case *syntax.Foreach:
		globFailureUse(ctx, x.Expr, message)
	case *syntax.FuncCall:
		// Cheap lexical filter first: almost no call is an array-only consumer.
		name, ok := x.Name.(*syntax.Name)
		if !ok {
			return
		}
		consumer, ok := globArrayConsumers[strings.ToLower(name.Value[strings.LastIndexByte(name.Value, '\\')+1:])]
		if ok && NativeBuiltinName(ctx, x) != "" {
			globFailureUse(ctx, CallArgument(x.Args, consumer.position, consumer.name), message)
		}
	}
}

// globArrayConsumers lists builtins whose array parameter rejects false: a
// TypeError since PHP 8.0, a warning and a meaningless result before.
var globArrayConsumers = map[string]struct {
	position int
	name     string
}{
	"count": {0, "value"}, "sizeof": {0, "value"},
	"array_map": {1, "array"}, "array_filter": {0, "array"}, "array_values": {0, "array"},
	"array_keys": {0, "array"}, "array_reverse": {0, "array"}, "array_slice": {0, "array"},
	"array_unique": {0, "array"}, "array_flip": {0, "array"}, "array_sum": {0, "array"},
	"array_product": {0, "array"}, "array_column": {0, "array"}, "array_chunk": {0, "array"},
	"array_key_first": {0, "array"}, "array_key_last": {0, "array"}, "array_is_list": {0, "array"},
	"array_rand": {0, "array"}, "array_key_exists": {1, "array"}, "in_array": {1, "haystack"},
	"array_search": {1, "haystack"}, "implode": {1, "array"}, "join": {1, "array"},
	"array_combine": {0, "keys"}, "array_fill_keys": {0, "keys"},
	// Variadic array lists cannot be named; only the first array is inspected.
	"array_merge": {0, ""}, "array_diff": {0, ""}, "array_intersect": {0, ""},
}

// globFailureUse reports an array-only use whose single reaching value is a
// builtin glob() call that no flow fact or recognised guard separates from false.
func globFailureUse(ctx *analysis.Context, use syntax.Expr, message string) {
	call, ok := NativeLocalBuiltinResult(ctx, use, "false", "glob").(*syntax.FuncCall)
	if !ok || NativeBuiltinName(ctx, call) != "glob" || ctx.Flow().Excludes(use, "false") || !NativeFailureFlowOnly(ctx, use) && globFailureGuarded(ctx, use) {
		return
	}
	ctx.ReportNode(use, message)
}

// globFailureGuarded proves that a dominating condition on the same reaching
// value already excluded false: an enclosing branch, a short-circuit operand,
// a ternary arm, or an earlier sibling `if` whose body always leaves.
func globFailureGuarded(ctx *analysis.Context, use syntax.Expr) bool {
	if _, ok := syntax.UnwrapParens(use).(*syntax.Variable); !ok {
		return false
	}
	span := use.Span()
	for p := use.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		switch x := p.(type) {
		case *syntax.Binary:
			and := x.Op.Kind == syntax.TBooleanAnd || x.Op.Kind == syntax.TAnd
			or := x.Op.Kind == syntax.TBooleanOr || x.Op.Kind == syntax.TOr
			if (and || or) && x.Right.Span().Contains(span) && globFailureExcluded(ctx, use, x.Left, and) {
				return true
			}
		case *syntax.Ternary:
			if x.Then != nil && x.Then.Span().Contains(span) && globFailureExcluded(ctx, use, x.Cond, true) {
				return true
			}
			if x.Else != nil && x.Else.Span().Contains(span) && globFailureExcluded(ctx, use, x.Cond, false) {
				return true
			}
		case *syntax.If:
			// Outside the body, every elseif condition, elseif body and the else
			// body run only after the leading condition failed.
			inBody := x.Body != nil && x.Body.Span().Contains(span)
			if !x.Cond.Span().Contains(span) && globFailureExcluded(ctx, use, x.Cond, inBody) {
				return true
			}
			for _, branch := range x.ElseIfs {
				if branch.Body != nil && branch.Body.Span().Contains(span) && globFailureExcluded(ctx, use, branch.Cond, true) {
					return true
				}
			}
		}
		st, ok := p.(syntax.Stmt)
		if !ok {
			continue
		}
		list, i, _ := astquery.StmtList(ctx.File, st)
		for _, previous := range list[:max(i, 0)] {
			if guard, ok := previous.(*syntax.If); ok && syntax.Terminates(guard.Body) && globFailureExcluded(ctx, use, guard.Cond, false) {
				return true
			}
			if globFailureLeaves(ctx, use, previous) {
				return true
			}
		}
	}
	return false
}

// globFailureLeaves recognises `$paths or exit;` and `$paths === false and
// throw ...;` statements: execution continues only when the left operand
// excluded false.
func globFailureLeaves(ctx *analysis.Context, use syntax.Expr, st syntax.Stmt) bool {
	statement, ok := st.(*syntax.ExprStmt)
	if !ok {
		return false
	}
	b, ok := syntax.UnwrapParens(statement.Expr).(*syntax.Binary)
	if !ok {
		return false
	}
	switch syntax.UnwrapParens(b.Right).(type) {
	case *syntax.Exit, *syntax.Throw:
	default:
		return false
	}
	switch b.Op.Kind {
	case syntax.TBooleanOr, syntax.TOr:
		return globFailureExcluded(ctx, use, b.Left, true)
	case syntax.TBooleanAnd, syntax.TAnd:
		return globFailureExcluded(ctx, use, b.Left, false)
	}
	return false
}

// globFailureExcluded reports whether condition evaluating to truth proves the
// guarded value is not false. glob() yields an array or false, so truthiness,
// emptiness and type predicates all separate the two outcomes.
func globFailureExcluded(ctx *analysis.Context, use, condition syntax.Expr, truth bool) bool {
	switch c := syntax.UnwrapParens(condition).(type) {
	case *syntax.Variable:
		return truth && nativeGuardSameValue(ctx, use, c)
	case *syntax.Empty:
		return !truth && nativeGuardSameValue(ctx, use, c.Expr)
	case *syntax.Isset:
		// A set offset means the value is an array; isset($paths) alone only
		// excludes null.
		for _, v := range c.Vars {
			if fetch, ok := syntax.UnwrapParens(v).(*syntax.ArrayDimFetch); ok && truth && nativeGuardSameValue(ctx, use, fetch.Var) {
				return true
			}
		}
	case *syntax.Unary:
		return c.Op.Kind == syntax.TExclaim && globFailureExcluded(ctx, use, c.Expr, !truth)
	case *syntax.FuncCall:
		switch NativeBuiltinName(ctx, c) {
		case "is_array", "is_iterable", "is_countable":
			return truth && nativeGuardSameValue(ctx, use, CallArgument(c.Args, 0, "value"))
		case "is_bool":
			return !truth && nativeGuardSameValue(ctx, use, CallArgument(c.Args, 0, "value"))
		}
	case *syntax.Binary:
		switch c.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd:
			return truth && (globFailureExcluded(ctx, use, c.Left, true) || globFailureExcluded(ctx, use, c.Right, true))
		case syntax.TBooleanOr, syntax.TOr:
			return !truth && (globFailureExcluded(ctx, use, c.Left, false) || globFailureExcluded(ctx, use, c.Right, false))
		case syntax.TIsIdentical, syntax.TIsEqual, syntax.TIsNotIdentical, syntax.TIsNotEqual:
			// Comparing with a boolean literal keeps or flips the operand's truth.
			negated := c.Op.Kind == syntax.TIsNotIdentical || c.Op.Kind == syntax.TIsNotEqual
			for _, pair := range [][2]syntax.Expr{{c.Left, c.Right}, {c.Right, c.Left}} {
				constant, ok := syntax.UnwrapParens(pair[1]).(*syntax.ConstFetch)
				if !ok || constant.Name == nil {
					continue
				}
				switch strings.ToLower(strings.TrimPrefix(constant.Name.Value, `\`)) {
				case "true":
					return globFailureExcluded(ctx, use, pair[0], truth != negated)
				case "false":
					return globFailureExcluded(ctx, use, pair[0], truth == negated)
				}
			}
		}
	}
	return false
}

// CheckOwnedStreamNotClosed checks the independently specified contract.
func CheckOwnedStreamNotClosed(ctx *analysis.Context, n syntax.Node, message string) {
	ret, ok := n.(*syntax.Return)
	if !ok {
		return
	}

	for _, call := range ctx.Flow().Calls(syntax.EnclosingVariableScope(ret)) {
		open, ok := call.Node.(*syntax.FuncCall)
		if !ok || call.Name != "fopen" || open.Span().End >= ret.Span().Start {
			continue
		}
		assignment, ok := open.Parent().(*syntax.Assign)
		if !ok {
			continue
		}
		variable, ok := assignment.Var.(*syntax.Variable)
		if !ok {
			continue
		}
		var use syntax.Expr
		if ret.Expr != nil {
			syntax.Inspect(ret.Expr, func(child syntax.Node) bool {
				if v, ok := child.(*syntax.Variable); ok && v.Name == variable.Name {
					use = v
				}
				return true
			})
		}
		if use == nil {
			continue
		}
		if _, ok := ret.Expr.(*syntax.Variable); ok {
			continue
		}
		state, known := ctx.Flow().StateBefore(ret, use)
		if !known || state.Operation == "fclose" || nativeFinally(ctx, ret, use, "fclose") {
			continue
		}
		if !ctx.Flow().Excludes(use, "false") {
			continue
		}
		if ownedStreamReleased(ctx, syntax.EnclosingVariableScope(ret), ret, variable.Name) {
			continue
		}
		ctx.ReportNode(ret, message)
		return
	}
}

// ownedStreamReleased reports whether the handle held by the named local may
// be closed by the return expression itself or may outlive the scope: any use
// other than a test, a rebinding or a borrowing builtin stream call hands the
// handle to code that can keep or close it.
func ownedStreamReleased(ctx *analysis.Context, scope syntax.Node, ret *syntax.Return, name string) bool {
	released := false
	syntax.Inspect(scope, func(n syntax.Node) bool {
		if released {
			return false
		}
		switch n := n.(type) {
		case *syntax.Closure:
			if n == scope {
				return true
			}
			for _, captured := range n.Uses {
				released = released || captured.Var.Name == name
			}
			return false
		case *syntax.ArrowFunction:
			// An arrow function captures every outer local its body reads.
			syntax.Inspect(n.Expr, func(inner syntax.Node) bool {
				v, ok := inner.(*syntax.Variable)
				released = released || (ok && v.Name == name)
				return true
			})
			return false
		case *syntax.Variable:
			if n.Name != name {
				return true
			}
			switch ownedStreamUse(ctx, n) {
			case ownedStreamEscapes:
				released = true
			case ownedStreamCloses:
				released = n.Span().Start >= ret.Span().Start && n.Span().End <= ret.Span().End
			}
		}
		return n == scope || !syntax.IsVariableScope(n)
	})
	return released
}

const (
	ownedStreamBorrowed = iota
	ownedStreamCloses
	ownedStreamEscapes
)

// ownedStreamUse classifies one read or write of the local holding a handle.
func ownedStreamUse(ctx *analysis.Context, v *syntax.Variable) int {
	var child syntax.Node = v
	parent := v.Parent()
	for {
		paren, ok := parent.(*syntax.Paren)
		if !ok {
			break
		}
		child, parent = paren, paren.Parent()
	}
	switch p := parent.(type) {
	case *syntax.Assign:
		// Rebinding the local is not a use of the handle; copying it is.
		if p.Var == child && !p.ByRef {
			return ownedStreamBorrowed
		}
	case *syntax.Binary:
		switch p.Op.Kind {
		case syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TIsEqual, syntax.TIsNotEqual,
			syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr:
			return ownedStreamBorrowed
		}
	case *syntax.Unary:
		if p.Op.Kind == syntax.TExclaim {
			return ownedStreamBorrowed
		}
	case *syntax.If, *syntax.ElseIf, *syntax.While, *syntax.DoWhile:
		return ownedStreamBorrowed
	case *syntax.Ternary:
		// A full ternary only tests its condition; `$h ?: x` yields the handle.
		if p.Cond == child && p.Then != nil {
			return ownedStreamBorrowed
		}
	case *syntax.Arg:
		call, ok := p.Parent().Parent().(*syntax.FuncCall)
		if !ok || p.ByRef || p.Unpack {
			return ownedStreamEscapes
		}
		switch NativeBuiltinName(ctx, call) {
		case "fclose":
			return ownedStreamCloses
		case "fread", "fwrite", "fputs", "fgets", "fgetc", "fgetcsv", "fputcsv", "fscanf", "fprintf", "vfprintf",
			"fpassthru", "feof", "fflush", "fsync", "fdatasync", "fseek", "ftell", "rewind", "ftruncate", "fstat", "flock",
			"stream_get_contents", "stream_get_line", "stream_get_meta_data", "stream_copy_to_stream",
			"stream_set_blocking", "stream_set_timeout", "stream_set_chunk_size", "stream_set_read_buffer",
			"stream_set_write_buffer", "stream_isatty", "stream_is_local", "stream_supports_lock",
			"is_resource", "get_resource_type", "get_resource_id":
			return ownedStreamBorrowed
		}
	}
	return ownedStreamEscapes
}

// CheckPathContainmentPrefixBoundary checks the independently specified contract.
func CheckPathContainmentPrefixBoundary(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.FuncCall)
	if !ok {
		return
	}
	// Each accepted form yields the tested path, the prefix expression and the
	// node to highlight; sized forms must compare exactly the prefix length.
	var subject, needle, length syntax.Expr
	var site syntax.Node = c
	sized := false
	switch NativeBuiltinName(ctx, c) {
	case "str_starts_with":
		subject, needle = CallArgument(c.Args, 0, "haystack"), CallArgument(c.Args, 1, "needle")
	case "strpos":
		// Only an identity test against 0 is a prefix test: a loose one also
		// accepts a missing needle.
		cmp, other, strict := nativePathPrefixComparison(c)
		offset := CallArgument(c.Args, 2, "offset")
		if !strict || !nativePathPrefixZero(ctx, other) || (offset != nil && !nativePathPrefixZero(ctx, offset)) {
			return
		}
		subject, needle, site = CallArgument(c.Args, 0, "haystack"), CallArgument(c.Args, 1, "needle"), cmp
	case "strncmp":
		cmp, other, _ := nativePathPrefixComparison(c)
		if !nativePathPrefixZero(ctx, other) {
			return
		}
		subject, needle, site = CallArgument(c.Args, 0, "string1"), CallArgument(c.Args, 1, "string2"), cmp
		if _, literal := NativeString(ctx, subject); literal {
			subject, needle = needle, subject
		}
		length, sized = CallArgument(c.Args, 2, "length"), true
	case "substr":
		cmp, other, _ := nativePathPrefixComparison(c)
		if cmp == nil || !nativePathPrefixZero(ctx, CallArgument(c.Args, 1, "offset")) {
			return
		}
		subject, needle, site = CallArgument(c.Args, 0, "string"), other, cmp
		length, sized = CallArgument(c.Args, 2, "length"), true
	default:
		return
	}
	prefix, known := NativeString(ctx, needle)
	if !known || prefix == "" || strings.HasSuffix(prefix, "/") || strings.HasSuffix(prefix, `\`) {
		return
	}
	if sized && !nativePathPrefixLength(ctx, length, prefix) {
		return
	}
	origin, ok := NativeValue(ctx, subject).(*syntax.FuncCall)
	if ok && NativeBuiltinName(ctx, origin) == "realpath" {
		ctx.ReportNode(site, message)
	}
}

// nativePathPrefixComparison returns the equality or identity comparison whose
// operand is the call, the opposite operand and whether the test is strict.
func nativePathPrefixComparison(c *syntax.FuncCall) (cmp *syntax.Binary, other syntax.Expr, strict bool) {
	var operand syntax.Expr = c
	for {
		paren, ok := operand.Parent().(*syntax.Paren)
		if !ok {
			break
		}
		operand = paren
	}
	cmp, ok := operand.Parent().(*syntax.Binary)
	if !ok {
		return nil, nil, false
	}
	switch cmp.Op.Kind {
	case syntax.TIsIdentical, syntax.TIsNotIdentical:
		strict = true
	case syntax.TIsEqual, syntax.TIsNotEqual:
	default:
		return nil, nil, false
	}
	other = cmp.Right
	if operand == cmp.Right {
		other = cmp.Left
	}
	return cmp, other, strict
}

func nativePathPrefixZero(ctx *analysis.Context, e syntax.Expr) bool {
	n, known := NativeInt(ctx, e)
	return known && n == 0
}

// nativePathPrefixLength proves that a length argument is the byte length of
// the prefix, written as an integer or as strlen() of the same text.
func nativePathPrefixLength(ctx *analysis.Context, e syntax.Expr, prefix string) bool {
	if n, known := NativeInt(ctx, e); known {
		return n == int64(len(prefix))
	}
	measured, ok := NativeValue(ctx, e).(*syntax.FuncCall)
	if !ok || NativeBuiltinName(ctx, measured) != "strlen" {
		return false
	}
	text, known := NativeString(ctx, CallArgument(measured.Args, 0, "string"))
	return known && text == prefix
}

// CheckReadModifyWriteLockTooLate checks the independently specified contract.
func CheckReadModifyWriteLockTooLate(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.FuncCall)
	if !ok {
		return
	}
	name := NativeBuiltinName(ctx, c)

	if name != "file_put_contents" {
		return
	}
	flags := CallArgument(c.Args, 2, "flags")
	present, known := NativeFlagContains(ctx, flags, "LOCK_EX")
	if !known || !present {
		return
	}
	data := CallArgument(c.Args, 1, "data")
	if data == nil {
		return
	}
	path := CallArgument(c.Args, 0, "filename")
	// The earliest matching read is the one a spanning lock has to precede.
	var first *syntax.FuncCall
	syntax.Inspect(data, func(child syntax.Node) bool {
		expr, ok := child.(syntax.Expr)
		if !ok {
			return true
		}
		value := NativeValue(ctx, expr)
		for {
			unary, ok := value.(*syntax.Unary)
			if !ok || !unary.Op.Kind.IsCast() {
				break
			}
			value = unary.Expr
		}
		if read, ok := value.(*syntax.FuncCall); ok && NativeBuiltinName(ctx, read) == "file_get_contents" && nativeSameExpression(ctx, path, CallArgument(read.Args, 0, "filename")) {
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
			name := NativeBuiltinName(ctx, call)
			handle := CallArgument(call.Args, 0, "stream")
			switch {
			case name == "":
				delegated = delegated || readModifyWriteLockDelegated(ctx, call.Name, call.Args, read, path)
			case handle == nil || (name != "flock" && name != "fclose"):
			default:
				key := string(ctx.File.Src[handle.Span().Start:handle.Span().End])
				// LOCK_UN (3) shares a bit with LOCK_EX (2), so compare the
				// whole operation once the optional LOCK_NB (4) is masked out.
				operation, known := fileLockOperation(ctx, CallArgument(call.Args, 1, "operation"))
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
	return nativeLockFollows(statement, write) || (list && nativeLockWithin(lock, owner))
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
		if arg, ok := node.(*syntax.Arg); ok && nativeSameExpression(ctx, path, arg.Value) {
			return true
		}
	}
	return false
}

// CheckStreamOpenFailureUnchecked checks the independently specified contract.
func CheckStreamOpenFailureUnchecked(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.FuncCall)
	if !ok {
		return
	}
	name := NativeBuiltinName(ctx, c)
	report := false

	switch name {
	case "fread", "fwrite", "fgets", "fseek", "ftell", "fflush", "ftruncate", "rewind":
	default:
		return
	}
	handle := CallArgument(c.Args, 0, "stream")
	if handle == nil {
		return
	}

	open, ok := NativeLocalResultOf(ctx, handle, "false", func(result syntax.Expr) bool {
		call, ok := result.(*syntax.FuncCall)
		return ok && NativeBuiltin(ctx, call, "fopen") && (!NativeFailureFlowOnly(ctx, handle) || !nativeJoinedWarningUnknown(ctx, call))
	}).(*syntax.FuncCall)
	report = ok && NativeBuiltinName(ctx, open) == "fopen" && !ctx.Flow().Excludes(handle, "false") && (NativeFailureFlowOnly(ctx, handle) || !streamOpenGuarded(ctx, handle, open))

	if report {
		ctx.ReportNode(c, message)
	}
}

// streamOpenGuarded proves that a dominating condition on the same reaching
// fopen() result already excluded false: an enclosing branch or loop
// condition, a short-circuit operand, a ternary arm, or an earlier sibling
// statement that leaves when the open failed.
func streamOpenGuarded(ctx *analysis.Context, use syntax.Expr, open *syntax.FuncCall) bool {
	if _, ok := syntax.UnwrapParens(use).(*syntax.Variable); !ok {
		return false
	}
	span := use.Span()
	for p := use.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		switch x := p.(type) {
		case *syntax.Binary:
			and := x.Op.Kind == syntax.TBooleanAnd || x.Op.Kind == syntax.TAnd
			or := x.Op.Kind == syntax.TBooleanOr || x.Op.Kind == syntax.TOr
			if (and || or) && x.Right.Span().Contains(span) && streamOpenExcluded(ctx, use, open, x.Left, and) {
				return true
			}
		case *syntax.Ternary:
			if x.Then != nil && x.Then.Span().Contains(span) && streamOpenExcluded(ctx, use, open, x.Cond, true) {
				return true
			}
			if x.Else != nil && x.Else.Span().Contains(span) && streamOpenExcluded(ctx, use, open, x.Cond, false) {
				return true
			}
		case *syntax.While:
			if x.Body != nil && x.Body.Span().Contains(span) && streamOpenExcluded(ctx, use, open, x.Cond, true) {
				return true
			}
		case *syntax.If:
			// Outside the body, every elseif condition, elseif body and the else
			// body run only after the leading condition failed.
			inBody := x.Body != nil && x.Body.Span().Contains(span)
			if !x.Cond.Span().Contains(span) && streamOpenExcluded(ctx, use, open, x.Cond, inBody) {
				return true
			}
			for _, branch := range x.ElseIfs {
				if branch.Body != nil && branch.Body.Span().Contains(span) && streamOpenExcluded(ctx, use, open, branch.Cond, true) {
					return true
				}
			}
		}
		st, ok := p.(syntax.Stmt)
		if !ok {
			continue
		}
		list, i, _ := astquery.StmtList(ctx.File, st)
		for _, previous := range list[:max(i, 0)] {
			if guard, ok := previous.(*syntax.If); ok && syntax.Terminates(guard.Body) && streamOpenExcluded(ctx, use, open, guard.Cond, false) {
				return true
			}
			if streamOpenLeaves(ctx, use, open, previous) {
				return true
			}
		}
	}
	return false
}

// streamOpenLeaves recognises expression statements that only complete when
// the open succeeded: `$h = fopen(...) or die();`, `$h === false and throw
// ...;` and `$h ?: throw ...;`.
func streamOpenLeaves(ctx *analysis.Context, use syntax.Expr, open *syntax.FuncCall, st syntax.Stmt) bool {
	statement, ok := st.(*syntax.ExprStmt)
	if !ok {
		return false
	}
	leaves := func(e syntax.Expr) bool {
		_, throws := syntax.UnwrapParens(e).(*syntax.Throw)
		return e != nil && (throws || syntax.ExitInvocation(e))
	}
	switch x := syntax.UnwrapParens(statement.Expr).(type) {
	case *syntax.Binary:
		if !leaves(x.Right) {
			return false
		}
		switch x.Op.Kind {
		case syntax.TBooleanOr, syntax.TOr:
			return streamOpenExcluded(ctx, use, open, x.Left, true)
		case syntax.TBooleanAnd, syntax.TAnd:
			return streamOpenExcluded(ctx, use, open, x.Left, false)
		}
	case *syntax.Ternary:
		if leaves(x.Else) {
			return streamOpenExcluded(ctx, use, open, x.Cond, true)
		}
		return leaves(x.Then) && streamOpenExcluded(ctx, use, open, x.Cond, false)
	}
	return false
}

// streamOpenExcluded reports whether condition evaluating to truth proves the
// handle is not false. fopen() yields a resource or false, so truthiness,
// emptiness and type predicates all separate the two outcomes.
func streamOpenExcluded(ctx *analysis.Context, use syntax.Expr, open *syntax.FuncCall, condition syntax.Expr, truth bool) bool {
	switch c := syntax.UnwrapParens(condition).(type) {
	case *syntax.Variable:
		return truth && nativeGuardSameValue(ctx, use, c)
	case *syntax.Assign:
		// `if ($h = fopen(...))` tests the very call whose result reaches the use.
		return truth && c.Op.Kind == syntax.TEqual && !c.ByRef && NativeUnsilenced(c.Value) == syntax.Expr(open) && astquery.Equivalent(ctx.File, use, c.Var)
	case *syntax.Empty:
		return !truth && nativeGuardSameValue(ctx, use, c.Expr)
	case *syntax.Unary:
		return c.Op.Kind == syntax.TExclaim && streamOpenExcluded(ctx, use, open, c.Expr, !truth)
	case *syntax.FuncCall:
		switch NativeBuiltinName(ctx, c) {
		case "is_resource":
			return truth && nativeGuardSameValue(ctx, use, CallArgument(c.Args, 0, "value"))
		case "is_bool":
			return !truth && nativeGuardSameValue(ctx, use, CallArgument(c.Args, 0, "value"))
		}
	case *syntax.Binary:
		switch c.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd:
			return truth && (streamOpenExcluded(ctx, use, open, c.Left, true) || streamOpenExcluded(ctx, use, open, c.Right, true))
		case syntax.TBooleanOr, syntax.TOr:
			return !truth && (streamOpenExcluded(ctx, use, open, c.Left, false) || streamOpenExcluded(ctx, use, open, c.Right, false))
		case syntax.TIsIdentical, syntax.TIsEqual, syntax.TIsNotIdentical, syntax.TIsNotEqual:
			// Comparing with a boolean literal keeps or flips the operand's truth.
			negated := c.Op.Kind == syntax.TIsNotIdentical || c.Op.Kind == syntax.TIsNotEqual
			for _, pair := range [][2]syntax.Expr{{c.Left, c.Right}, {c.Right, c.Left}} {
				constant, ok := syntax.UnwrapParens(pair[1]).(*syntax.ConstFetch)
				if !ok || constant.Name == nil {
					continue
				}
				switch strings.ToLower(strings.TrimPrefix(constant.Name.Value, `\`)) {
				case "true":
					return streamOpenExcluded(ctx, use, open, pair[0], truth != negated)
				case "false":
					return streamOpenExcluded(ctx, use, open, pair[0], truth == negated)
				}
			}
		}
	}
	return false
}

// CheckStreamUseAfterClose checks the independently specified contract.
func CheckStreamUseAfterClose(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.FuncCall)
	if !ok {
		return
	}
	name := NativeBuiltinName(ctx, c)

	switch name {
	case "fread", "fwrite", "fgets", "fseek", "ftell", "fflush", "ftruncate", "rewind":
	default:
		return
	}
	handle := CallArgument(c.Args, 0, "stream")
	if handle == nil {
		return
	}
	state, closed := ctx.Flow().StateBefore(c, handle, "fclose")
	if !closed {
		return
	}
	scope := syntax.EnclosingVariableScope(c)
	if streamCapturedByReference(ctx, scope, handle) || streamCloseFailed(ctx, state.Span, c, scope) {
		return
	}
	ctx.ReportNode(c, message)
}

// streamInspectScope walks the nodes of one variable scope, nested scopes
// included; nil denotes the file-level scope.
func streamInspectScope(ctx *analysis.Context, scope syntax.Node, visit func(syntax.Node) bool) {
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
		return
	}
	syntax.Inspect(scope, visit)
}

// streamCapturedByReference reports whether a closure created in scope binds
// the handle variable by reference: calling it can store a freshly opened
// stream in the variable, which the flow state of the closed one cannot see.
func streamCapturedByReference(ctx *analysis.Context, scope syntax.Node, handle syntax.Expr) bool {
	variable, _ := syntax.UnwrapParens(handle).(*syntax.Variable)
	captured := false
	streamInspectScope(ctx, scope, func(n syntax.Node) bool {
		if n == scope {
			return true
		}
		if closure, ok := n.(*syntax.Closure); ok {
			for _, use := range closure.Uses {
				captured = captured || (variable != nil && use.ByRef && use.Var.Name == variable.Name)
			}
		}
		return !captured && !syntax.IsVariableScope(n)
	})
	return captured
}

// streamCloseFailed proves that use runs only when the fclose call spanning
// closed returned false, through a test of the call itself or of an unchanged
// local holding its result. A refused close leaves the stream open.
func streamCloseFailed(ctx *analysis.Context, closed syntax.Span, use *syntax.FuncCall, scope syntax.Node) bool {
	guarded := false
	for _, reached := range ctx.Flow().Calls(scope) {
		call, ok := reached.Node.(syntax.Expr)
		if !ok || call.Span() != closed {
			continue
		}
		stored, failed, direct := nativeLockTruthGuards(call, nativeLockTruth{whenFalse: true}, use)
		if guarded = direct; guarded || stored == nil {
			continue
		}
		name := stored.Var.(*syntax.Variable).Name
		value := ctx.Flow().Value(stored.Value).Expr
		streamInspectScope(ctx, scope, func(n syntax.Node) bool {
			if guarded || (n != scope && syntax.IsVariableScope(n)) {
				return false
			}
			// Only a read whose single reaching value is the stored result
			// tests this close; a reassigned or merged local proves nothing.
			if read, ok := n.(*syntax.Variable); ok && read.Name == name && value != nil && ctx.Flow().Value(read).Expr == value {
				_, _, guarded = nativeLockTruthGuards(read, failed, use)
			}
			return true
		})
	}
	return guarded
}

// CheckTempnamDirectoryFallbackUnchecked checks the independently specified contract.
func CheckTempnamDirectoryFallbackUnchecked(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.FuncCall)
	if !ok {
		return
	}
	name := NativeBuiltinName(ctx, c)
	report := false

	if name != "file_put_contents" && name != "fopen" {
		return
	}
	path := CallArgument(c.Args, 0, "filename")
	open, ok := NativeValue(ctx, path).(*syntax.FuncCall)
	if !ok || NativeBuiltinName(ctx, open) != "tempnam" {
		return
	}
	// Asking for the system temporary directory leaves nothing to fall back to.
	directory := CallArgument(open.Args, 0, "directory")
	if value := NativeValue(ctx, directory); value != nil {
		directory = value
	}
	if system, ok := syntax.UnwrapParens(directory).(*syntax.FuncCall); ok && NativeBuiltinName(ctx, system) == "sys_get_temp_dir" {
		return
	}
	report = !nativeDirectoryGuard(ctx, c, path, open)

	if report {
		ctx.ReportNode(c, message)
	}
}

// nativeIteratorReassigned suppresses a lexical foreach contract after writes
// that can change the element binding before its destructive use.
func nativeIteratorReassigned(loop *syntax.Foreach, at syntax.Node, name string) bool {
	changed := false
	syntax.Inspect(loop.Body, func(node syntax.Node) bool {
		if syntax.IsVariableScope(node) {
			return false
		}
		if node.Span().Start >= at.Span().Start {
			return false
		}
		var targets []syntax.Expr
		switch n := node.(type) {
		case *syntax.Assign:
			targets = []syntax.Expr{n.Var}
		case *syntax.IncDec:
			targets = []syntax.Expr{n.Var}
		case *syntax.Unset:
			targets = n.Vars
		case *syntax.Foreach:
			targets = []syntax.Expr{n.Value, n.Key}
		}
		for _, target := range targets {
			if variable, ok := target.(*syntax.Variable); ok && variable.Name == name {
				changed = true
			}
		}
		return true
	})
	return changed
}
