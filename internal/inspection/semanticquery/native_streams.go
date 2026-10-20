package semanticquery

import (
	"sort"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/types"
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

// nativeDirectoryVerified proves that the canonical parent of the temporary file
// held by path was compared with a canonical directory after source created it
// and that the comparison is known to have matched where at runs: an equality
// that the conditions guarding at prove true, or an inequality they prove
// false, while path still holds the same file.
func nativeDirectoryVerified(ctx *analysis.Context, at, path syntax.Expr, source syntax.Node) bool {
	variable, ok := syntax.UnwrapParens(path).(*syntax.Variable)
	if !ok || NativeLabelled(ctx, at) {
		return false
	}
	env := ctx.Types().Native()
	return NativeEarlierCallProves(ctx, at, func(n syntax.Expr) (types.Type, func(types.Type) bool) {
		cmp, ok := n.(*syntax.Binary)
		if !ok || cmp.Span().Start < source.Span().End {
			return types.Unknown, nil
		}
		// An equality must have held, an inequality failed. Canonical paths
		// are absolute strings or false: the loose operators decide alike.
		want := ""
		switch cmp.Op.Kind {
		case syntax.TIsIdentical, syntax.TIsEqual:
			want = "true"
		case syntax.TIsNotIdentical, syntax.TIsNotEqual:
			want = "false"
		default:
			return types.Unknown, nil
		}
		parent := nativeCanonicalParent(ctx, cmp.Left, variable.Name)
		if parent == nil || nativeCanonicalPath(ctx, cmp.Right) == nil {
			if parent = nativeCanonicalParent(ctx, cmp.Right, variable.Name); parent == nil || nativeCanonicalPath(ctx, cmp.Left) == nil {
				return types.Unknown, nil
			}
		}
		if !env.ReachesUnchanged(parent, at) {
			return types.Unknown, nil // another file since
		}
		return types.Bool, func(t types.Type) bool { return t.OnlyOf(want) }
	})
}

// nativeCanonicalPath returns the argument of a builtin realpath() call.
func nativeCanonicalPath(ctx *analysis.Context, e syntax.Expr) syntax.Expr {
	call, ok := syntax.UnwrapParens(e).(*syntax.FuncCall)
	if !ok || NativeBuiltinName(ctx, call) != "realpath" {
		return nil
	}
	return CallArgument(call.Args, 0, "path")
}

// nativeCanonicalParent returns the variable read by realpath(dirname($name)),
// nil for anything else.
func nativeCanonicalParent(ctx *analysis.Context, e syntax.Expr, name string) *syntax.Variable {
	dir, ok := syntax.UnwrapParens(nativeCanonicalPath(ctx, e)).(*syntax.FuncCall)
	if !ok || NativeBuiltinName(ctx, dir) != "dirname" {
		return nil
	}
	v, ok := syntax.UnwrapParens(CallArgument(dir.Args, 0, "path")).(*syntax.Variable)
	if !ok || v.Name != name {
		return nil
	}
	return v
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
// no longer name when c runs: what the conditions guarding c prove of the
// entry's earlier isDot()/isDir()/isFile()/isLink() results and of its name
// (getFilename(), argument-less getBasename()), joined over those calls.
func nativeDotEntriesExcludedAt(ctx *analysis.Context, loop *syntax.Foreach, c *syntax.FuncCall, variable string) uint8 {
	if NativeLabelled(ctx, c) {
		return 0 // a jump may land between a test and the call
	}
	env := ctx.Types().Native()
	var mask uint8
	syntax.Inspect(loop.Body, func(n syntax.Node) bool {
		if n.Span().Start >= c.Span().Start || syntax.IsVariableScope(n) {
			return false
		}
		call, ok := n.(*syntax.MethodCall)
		if !ok || call.Span().End > c.Span().Start {
			return true
		}
		switch method := nativeEntryMethod(call, variable); {
		case method == "isdot" || method == "isdir":
			if t := env.CallResultAsAt(call, c, types.Bool); t.IsUnknown() || t.OnlyOf("false") {
				mask = nativeDotEntryNames
			}
		case method == "isfile" || method == "islink":
			if t := env.CallResultAsAt(call, c, types.Bool); t.IsUnknown() || t.OnlyOf("true") {
				mask = nativeDotEntryNames
			}
		case nativeEntryName(call, variable):
			t := env.CallResultAsAt(call, c, types.TrackedStrings())
			if t.StringExcludes(".") {
				mask |= 1
			}
			if t.StringExcludes("..") {
				mask |= 2
			}
		}
		return true
	})
	return mask
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

// nativeEntryName reports whether call reads the entry's own name, which is
// "." or ".." exactly for the dot entries.
func nativeEntryName(call *syntax.MethodCall, variable string) bool {
	if call.Args != nil && len(call.Args.Args) != 0 {
		return false
	}
	name := nativeEntryMethod(call, variable)
	return name == "getfilename" || name == "getbasename"
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
		if ok && lock.Span() == state.Span && NativeCallSucceeded(ctx, lock, c) {
			return
		}
	}
	ctx.ReportNode(c, message)
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
			if !ok || reached.Name != "flock" || !NativeBuiltin(ctx, lock, "flock") {
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
			operation, known := fileLockOperation(ctx, CallArgument(lock.Args, 1, "operation"))
			if !known || operation&^4 != 2 {
				continue
			}
			stream := CallArgument(lock.Args, 0, "stream")
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
		if candidate.invalidated != current.Invalidated || !NativeDominates(lock, write) {
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
	if !ok || NativeBuiltinName(ctx, call) != "glob" || ctx.Flow().Excludes(use, "false") || !NativeFailureFlowOnly(ctx, use) && NativeNarrowedExcludes(ctx, use, types.Of("array", "false"), "false") {
		return
	}
	ctx.ReportNode(use, message)
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

	var open *syntax.FuncCall
	for _, candidate := range NativeLocalBuiltinResults(ctx, handle, "false", "fopen") {
		if call, ok := candidate.(*syntax.FuncCall); ok && (!NativeFailureFlowOnly(ctx, handle) || !nativeJoinedWarningUnknown(ctx, call)) {
			open = call
			break
		}
	}
	ok = open != nil
	report = ok && NativeBuiltinName(ctx, open) == "fopen" && !ctx.Flow().Excludes(handle, "false") && (NativeFailureFlowOnly(ctx, handle) || !NativeNarrowedExcludes(ctx, handle, types.Of("resource", "false"), "false"))

	if report {
		ctx.ReportNode(c, message)
	}
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
// closed returned false, through the tests on that call's result guarding
// use. A refused close leaves the stream open.
func streamCloseFailed(ctx *analysis.Context, closed syntax.Span, use *syntax.FuncCall, scope syntax.Node) bool {
	if NativeLabelled(ctx, use) {
		return false
	}
	for _, reached := range ctx.Flow().Calls(scope) {
		call, ok := reached.Node.(*syntax.FuncCall)
		if !ok || call.Span() != closed {
			continue
		}
		if t := ctx.Types().Native().CallResultAt(call, use); !t.IsUnknown() && !nativeMayHold(t, "true") {
			return true
		}
	}
	return false
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
	report = !nativeDirectoryVerified(ctx, c, path, open)

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
