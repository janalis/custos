// Package untrustedshellcommand implements the native UntrustedShellCommand inspection.
package untrustedshellcommand

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/flow"
)

const message = "Keep untrusted input out of shell command syntax."

// maxDepth bounds the walk through nested expressions and reaching values.
const maxDepth = 16

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule { return rule{} }
func (rule) ID() string  { return "UntrustedShellCommand" }
func (rule) Semantic()   {}
func (rule) Flow()       {}
func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall, syntax.KMethodCall, syntax.KStaticCall, syntax.KInterpolatedString}
}

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	switch n := n.(type) {
	case *syntax.InterpolatedString:
		// The backtick operator runs its text through the shell (D1).
		if n.Backtick && injectable(ctx, n, 0) {
			ctx.ReportNode(n, message)
		}
	case *syntax.MethodCall, *syntax.StaticCall:
		if wrapperSink(ctx, n) {
			ctx.ReportNode(n, message)
		}
	case *syntax.FuncCall:
		if reportsCall(ctx, n) {
			ctx.ReportNode(n, message)
		}
	}
}

func reportsCall(ctx *analysis.Context, c *syntax.FuncCall) bool {
	name := semanticquery.NativeBuiltinName(ctx, c)
	switch name {
	case "system", "exec", "shell_exec", "passthru", "popen", "proc_open":
		command := semanticquery.CallArgument(c.Args, 0, "command")
		if name == "proc_open" && argumentVector(ctx, command) { // E1
			return false
		}
		return injectable(ctx, command, 0)
	}
	return wrapperSink(ctx, c)
}

// argumentVector proves the array form of proc_open, which bypasses the shell.
func argumentVector(ctx *analysis.Context, command syntax.Expr) bool {
	if command == nil {
		return false
	}
	if _, ok := semanticquery.NativeValue(ctx, command).(*syntax.Array); ok {
		return true
	}
	return ctx.TypeOf(command).IsArrayLike()
}

// wrapperSink reports a project wrapper whose summary forwards request data
// from this call to a shell sink. The summary merges whole argument values,
// so arguments proven harmless here (E1, E3) do not count as carriers.
func wrapperSink(ctx *analysis.Context, c syntax.Node) bool {
	reaches := false
	for _, sink := range ctx.Flow().Sinks(c) {
		if sink.Context != flow.Shell || !sink.Argument.Complete || sink.Argument.Safe&flow.Shell != 0 {
			continue
		}
		for _, source := range sink.Argument.Sources {
			reaches = reaches || source.Parameter < 0
		}
	}
	if !reaches {
		return false
	}
	harmless := false
	var args *syntax.ArgList
	switch c := c.(type) {
	case *syntax.FuncCall:
		args = c.Args
	case *syntax.MethodCall:
		args = c.Args
	case *syntax.StaticCall:
		args = c.Args
	}
	for _, node := range args.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok || !ctx.Flow().Tainted(arg.Value, flow.Shell) {
			continue
		}
		if injectable(ctx, arg.Value, 0) {
			return true
		}
		harmless = true
	}
	// Without a tainted argument the wrapper reads the request itself.
	return !harmless
}

// injectable reports whether request data can reach shell syntax through x.
// The flow engine joins the taint of every operand; this walk drops the
// operands that only select or measure a value and cannot carry text.
func injectable(ctx *analysis.Context, x syntax.Expr, depth int) bool {
	if !ctx.Flow().Tainted(x, flow.Shell) {
		return false
	}
	if depth >= maxDepth {
		return true
	}
	any := func(xs ...syntax.Expr) bool {
		for _, e := range xs {
			if injectable(ctx, e, depth+1) {
				return true
			}
		}
		return false
	}
	switch n := x.(type) {
	case *syntax.Paren:
		return any(n.Expr)
	case *syntax.Variable:
		// A single reaching assignment stands for the variable.
		if value, ok := ctx.Flow().Resolve(n); ok && value != syntax.Expr(n) && !aliased(ctx, n) {
			return any(value)
		}
	case *syntax.InterpolatedString:
		return any(n.Parts...)
	case *syntax.Match:
		// Only arm results become the value; the subject merely selects one.
		for _, arm := range n.Arms {
			if any(arm.Body) {
				return true
			}
		}
		return false
	case *syntax.Ternary:
		return any(n.Then, n.Else)
	case *syntax.Unary:
		// Sign operators yield a number or fail on arrays (E3).
		return n.Op.Kind != syntax.TPlus && n.Op.Kind != syntax.TMinus
	case *syntax.Binary:
		switch n.Op.Kind {
		case syntax.TDot, syntax.TCoalesce:
			return any(n.Left, n.Right)
		case syntax.TPlus:
			// Two arrays form a union that keeps their request-derived elements.
			return !scalar(ctx, n.Left) && !scalar(ctx, n.Right)
		}
	case *syntax.FuncCall:
		return !scalarCall(ctx, n)
	}
	return true
}

// scalar proves that x is never an array.
func scalar(ctx *analysis.Context, x syntax.Expr) bool {
	switch n := syntax.UnwrapParens(x).(type) {
	case *syntax.Literal:
		return true
	case *syntax.FuncCall:
		return scalarCall(ctx, n)
	}
	return false
}

// scalarCall proves a resolved builtin whose result is boolean, numeric or
// null on the configured PHP version (E3).
func scalarCall(ctx *analysis.Context, c *syntax.FuncCall) bool {
	return semanticquery.NativeBuiltinName(ctx, c) != "" &&
		ctx.TypeOf(c).OnlyOf("int", "float", "bool", "false", "true", "null")
}
