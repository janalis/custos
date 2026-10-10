package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
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

func nativeDirectoryGuard(ctx *analysis.Context, at syntax.Node, path syntax.Expr) bool {
	variable, ok := path.(*syntax.Variable)
	if !ok {
		return false
	}
	for p := at.Parent(); p != nil; p = p.Parent() {
		block, ok := p.(*syntax.Block)
		if !ok {
			continue
		}
		for _, st := range block.Stmts {
			branch, ok := st.(*syntax.If)
			if !ok || branch.Span().End > at.Span().Start || !syntax.Terminates(branch.Body) {
				continue
			}
			binary, ok := branch.Cond.(*syntax.Binary)
			if !ok || binary.Op.Kind != syntax.TIsNotIdentical {
				continue
			}
			left, ok := binary.Left.(*syntax.FuncCall)
			if !ok || NativeBuiltinName(ctx, left) != "realpath" {
				continue
			}
			dir, ok := CallArgument(left.Args, 0, "path").(*syntax.FuncCall)
			if !ok || NativeBuiltinName(ctx, dir) != "dirname" {
				continue
			}
			v, ok := CallArgument(dir.Args, 0, "path").(*syntax.Variable)
			if !ok || v.Name != variable.Name {
				continue
			}
			right, ok := binary.Right.(*syntax.FuncCall)
			if ok && NativeBuiltinName(ctx, right) == "realpath" {
				return true
			}
		}
	}
	return false
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
		guarded := false
		syntax.Inspect(loop.Body, func(child syntax.Node) bool {
			branch, ok := child.(*syntax.If)
			if !ok || branch.Span().End > c.Span().Start {
				return true
			}
			check, ok := branch.Cond.(*syntax.MethodCall)
			if !ok {
				return true
			}
			id, ok := check.Name.(*syntax.Identifier)
			receiver, rok := check.Var.(*syntax.Variable)
			if ok && rok && id.Value == "isDot" && receiver.Name == variable.Name && syntax.Terminates(branch.Body) {
				guarded = true
			}
			return true
		})
		for ancestor := c.Parent(); ancestor != nil && ancestor != loop; ancestor = ancestor.Parent() {
			branch, ok := ancestor.(*syntax.If)
			if !ok {
				continue
			}
			not, ok := branch.Cond.(*syntax.Unary)
			if !ok || not.Op.Kind != syntax.TExclaim {
				continue
			}
			check, ok := not.Expr.(*syntax.MethodCall)
			if !ok {
				continue
			}
			id, ok := check.Name.(*syntax.Identifier)
			receiver, rok := check.Var.(*syntax.Variable)
			if ok && rok && id.Value == "isDot" && receiver.Name == variable.Name {
				guarded = true
			}
		}
		report = !guarded
		break
	}

	if report {
		ctx.ReportNode(c, message)
	}
}

// CheckFileLockFailureUnchecked checks the independently specified contract.
func CheckFileLockFailureUnchecked(ctx *analysis.Context, n syntax.Node, message string) {
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

	if name != "fwrite" && name != "ftruncate" {
		return
	}
	state, known := ctx.Flow().StateBefore(c, handle, "flock")
	report = known && !state.Successful

	if report {
		ctx.ReportNode(c, message)
	}
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
		if ok && arg.Name == variable.Name && ctx.Flow().Value(arg).Identity == ctx.Flow().Value(c).Identity {
			report = true
		}
	}

	if report {
		ctx.ReportNode(c, message)
	}
}

// CheckGlobFailureUnchecked checks the independently specified contract.
func CheckGlobFailureUnchecked(ctx *analysis.Context, n syntax.Node, message string) {
	loop, ok := n.(*syntax.Foreach)
	if !ok {
		return
	}

	call, ok := NativeValue(ctx, loop.Expr).(*syntax.FuncCall)
	if ok && NativeBuiltinName(ctx, call) == "glob" && !ctx.Flow().Excludes(loop.Expr, "false") {
		ctx.ReportNode(loop.Expr, message)
	}
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
		ctx.ReportNode(ret, message)
		return
	}
}

// CheckPathContainmentPrefixBoundary checks the independently specified contract.
func CheckPathContainmentPrefixBoundary(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.FuncCall)
	if !ok {
		return
	}
	name := NativeBuiltinName(ctx, c)
	report := false

	if name != "str_starts_with" {
		return
	}
	prefix, known := NativeString(ctx, CallArgument(c.Args, 1, "needle"))
	if !known || prefix == "" || strings.HasSuffix(prefix, "/") || strings.HasSuffix(prefix, `\`) {
		return
	}
	origin, ok := NativeValue(ctx, CallArgument(c.Args, 0, "haystack")).(*syntax.FuncCall)
	report = ok && NativeBuiltinName(ctx, origin) == "realpath"

	if report {
		ctx.ReportNode(c, message)
	}
}

// CheckReadModifyWriteLockTooLate checks the independently specified contract.
func CheckReadModifyWriteLockTooLate(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.FuncCall)
	if !ok {
		return
	}
	name := NativeBuiltinName(ctx, c)
	report := false

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
		if read, ok := value.(*syntax.FuncCall); ok && NativeBuiltinName(ctx, read) == "file_get_contents" && nativeSameExpression(ctx, CallArgument(c.Args, 0, "filename"), CallArgument(read.Args, 0, "filename")) {
			report = true
		}
		return true
	})

	if report {
		ctx.ReportNode(c, message)
	}
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

	open, ok := NativeValue(ctx, handle).(*syntax.FuncCall)
	report = ok && NativeBuiltinName(ctx, open) == "fopen" && !ctx.Flow().Excludes(handle, "false")

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
	_, report = ctx.Flow().StateBefore(c, handle, "fclose")

	if report {
		ctx.ReportNode(c, message)
	}
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
	report = ok && NativeBuiltinName(ctx, open) == "tempnam" && !nativeDirectoryGuard(ctx, c, path)

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
