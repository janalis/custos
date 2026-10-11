// Package untrustedfilesystempath implements the native UntrustedFilesystemPath inspection.
package untrustedfilesystempath

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/flow"
)

const message = "Validate filesystem containment before using this path."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "UntrustedFilesystemPath" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if call, ok := n.(*syntax.FuncCall); ok && pathSinks[semanticquery.NativeBuiltinName(ctx, call)] &&
		pathProven(ctx, semanticquery.CallArgument(call.Args, 0, "filename")) {
		return
	}
	semanticquery.CheckUntrustedFilesystemPath(ctx, n, message)
}

// pathSinks are the builtins whose first argument the shared check inspects.
var pathSinks = map[string]bool{
	"readfile": true, "fopen": true, "file_get_contents": true,
	"file_put_contents": true, "unlink": true, "rmdir": true,
}

// pathProven reports whether every request-derived part of a path is either a
// finite server-chosen name or a canonical path checked against a trusted
// directory. Concatenation is safe only when each operand is.
func pathProven(ctx *analysis.Context, e syntax.Expr) bool {
	if e == nil {
		return false
	}
	e = syntax.UnwrapParens(e)
	if !ctx.Flow().Tainted(e, flow.Path) {
		return true
	}
	if concat, ok := e.(*syntax.Binary); ok && concat.Op.Kind == syntax.TDot {
		return pathProven(ctx, concat.Left) && pathProven(ctx, concat.Right)
	}
	origin, call := semanticquery.NativeValue(ctx, e).(*syntax.FuncCall)
	path, ok := e.(*syntax.Variable)
	return call && ok && semanticquery.NativeBuiltinName(ctx, origin) == "realpath" && contained(ctx, path, origin) && !rebindable(ctx, path, origin)
}

// rebindable reports a write that may change the variable after the
// assignment of origin while its resolved value stays origin: compound,
// destructuring, offset and reference assignments, increments, loop
// variables, by-reference arguments and captures, catch bindings, static
// declarations, extract(), variable variables and the global symbol table.
// A direct write that ends before origin is superseded by it.
func rebindable(ctx *analysis.Context, v *syntax.Variable, origin syntax.Expr) bool {
	found := false
	start := origin.Span().Start
	hit := func(e syntax.Expr) bool { return e != nil && e.Span().End > start && writes(e, v.Name) }
	visit := func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.Assign:
			own := x.Op.Kind == syntax.TEqual && !x.ByRef && syntax.UnwrapParens(x.Value) == origin
			found = found || hit(x.Var) && !own || x.ByRef && writes(x.Value, v.Name)
		case *syntax.IncDec:
			found = found || hit(x.Var)
		case *syntax.Foreach:
			found = found || hit(x.Key) || hit(x.Value)
		case *syntax.ArgList:
			for i, a := range x.Args {
				if arg, ok := a.(*syntax.Arg); ok && hit(arg.Value) && (arg.ByRef || bindsByRef(ctx, x, i)) {
					found = true
				}
			}
		case *syntax.ClosureUse:
			found = found || x.ByRef && x.Var.Name == v.Name
		case *syntax.Catch:
			found = found || x.Var != nil && x.Var.Name == v.Name
		case *syntax.StaticVar:
			found = found || x.Var.Name == v.Name
		case *syntax.FuncCall:
			found = found || strings.EqualFold(astquery.CallLastName(x), "extract")
		case *syntax.Variable:
			_, global := x.Parent().(*syntax.Global)
			found = found || x.Name == "GLOBALS" || global && x.Name == v.Name
		}
		return !found
	}
	if scope := syntax.EnclosingVariableScope(v); scope != nil {
		syntax.Inspect(scope, visit)
	} else {
		syntax.InspectFile(ctx.File, visit)
	}
	return found
}

// writes reports whether assigning to e may rebind or alter the named
// variable: the variable itself, one of its offsets, a destructuring slot
// holding either, or a variable variable.
func writes(e syntax.Expr, name string) bool {
	var items []*syntax.ArrayItem
	switch x := syntax.UnwrapParens(e).(type) {
	case *syntax.Variable:
		return x.NameExpr != nil || x.Name == name
	case *syntax.ArrayDimFetch:
		return writes(x.Var, name)
	case *syntax.Array:
		items = x.Items
	case *syntax.List:
		items = x.Items
	}
	for _, item := range items {
		if item != nil && writes(item.Value, name) {
			return true
		}
	}
	return false
}

// bindsByRef reports whether argument i of list binds to a by-reference
// parameter of a resolved callee (named, positional or variadic tail).
func bindsByRef(ctx *analysis.Context, list *syntax.ArgList, i int) bool {
	callee, _ := semanticquery.ResolveCallee(ctx, list.Parent())
	params := callee.Params
	if name := list.Args[i].(*syntax.Arg).Name; name != nil {
		for _, p := range params {
			if p.ByRef && strings.EqualFold(strings.TrimPrefix(p.Name, "$"), name.Value) {
				return true
			}
		}
		return false
	}
	if i >= len(params) {
		i = len(params) - 1
		if i < 0 || !params[i].Variadic {
			return false
		}
	}
	return params[i].ByRef
}

// contained looks for a prefix test on the same canonical value that every
// path to use has passed: an enclosing `if` body, or an earlier sibling `if`
// that leaves the scope when the test fails.
func contained(ctx *analysis.Context, use *syntax.Variable, origin syntax.Expr) bool {
	for n := syntax.Node(use); n != nil && !syntax.IsVariableScope(n); n = n.Parent() {
		if branch, ok := n.Parent().(*syntax.If); ok && n == syntax.Node(branch.Body) && implies(ctx, branch.Cond, true, origin) {
			return true
		}
		st, ok := n.(syntax.Stmt)
		if !ok {
			continue
		}
		list, idx, _ := astquery.StmtList(ctx.File, st)
		for i := 0; i < idx; i++ {
			guard, ok := list[i].(*syntax.If)
			if ok && syntax.Terminates(guard.Body) && implies(ctx, guard.Cond, false, origin) {
				return true
			}
		}
	}
	return false
}

// implies reports whether cond evaluating to truth proves the prefix test.
func implies(ctx *analysis.Context, cond syntax.Expr, truth bool, origin syntax.Expr) bool {
	switch x := syntax.UnwrapParens(cond).(type) {
	case *syntax.Unary:
		return x.Op.Kind == syntax.TExclaim && implies(ctx, x.Expr, !truth, origin)
	case *syntax.FuncCall:
		return truth && prefixTest(ctx, x, origin, "str_starts_with")
	case *syntax.Binary:
		switch x.Op.Kind {
		case syntax.TBooleanAnd:
			return truth && (implies(ctx, x.Left, true, origin) || implies(ctx, x.Right, true, origin))
		case syntax.TBooleanOr:
			return !truth && (implies(ctx, x.Left, false, origin) || implies(ctx, x.Right, false, origin))
		case syntax.TIsIdentical, syntax.TIsNotIdentical:
			if (x.Op.Kind == syntax.TIsIdentical) != truth {
				return false
			}
			for _, side := range [2][2]syntax.Expr{{x.Left, x.Right}, {x.Right, x.Left}} {
				call, ok := syntax.UnwrapParens(side[0]).(*syntax.FuncCall)
				if zero, known := semanticquery.NativeInt(ctx, side[1]); ok && known && zero == 0 {
					return prefixTest(ctx, call, origin, "strpos", "strncmp")
				}
			}
		}
	}
	return false
}

// prefixTest matches one of the named builtins comparing the canonical value
// against a trusted directory prefix from its first byte.
func prefixTest(ctx *analysis.Context, call *syntax.FuncCall, origin syntax.Expr, names ...string) bool {
	name := semanticquery.NativeBuiltinName(ctx, call)
	subject, prefix := "haystack", "needle"
	switch {
	case name == "" || name != names[0] && name != names[len(names)-1]:
		return false
	case name == "strncmp":
		subject, prefix = "string1", "string2"
	case semanticquery.CallArgument(call.Args, 2, "offset") != nil:
		return false
	}
	path, ok := syntax.UnwrapParens(semanticquery.CallArgument(call.Args, 0, subject)).(*syntax.Variable)
	if !ok || semanticquery.NativeValue(ctx, path) != origin {
		return false
	}
	size, literal, trusted := trustedPrefix(ctx, semanticquery.CallArgument(call.Args, 1, prefix))
	if !trusted || name != "strncmp" {
		return trusted
	}
	// A shorter comparison would also accept sibling directories.
	length, known := semanticquery.NativeInt(ctx, semanticquery.CallArgument(call.Args, 2, "length"))
	return literal && known && length >= int64(size)
}

// trustedPrefix accepts an absolute literal directory ending in a slash, or a
// server-chosen realpath() result proven not to be false followed by a
// separator. A false base would degrade to "/" and contain everything.
func trustedPrefix(ctx *analysis.Context, e syntax.Expr) (size int, literal, ok bool) {
	if s, known := semanticquery.NativeString(ctx, e); known {
		return len(s), true, len(s) > 1 && s[0] == '/' && s[len(s)-1] == '/'
	}
	concat, ok := syntax.UnwrapParens(e).(*syntax.Binary)
	if !ok || concat.Op.Kind != syntax.TDot {
		return 0, false, false
	}
	base, ok := syntax.UnwrapParens(concat.Left).(*syntax.Variable)
	if !ok {
		return 0, false, false
	}
	call, ok := semanticquery.NativeValue(ctx, base).(*syntax.FuncCall)
	if !ok || semanticquery.NativeBuiltinName(ctx, call) != "realpath" || !ctx.Flow().Excludes(base, "false") || rebindable(ctx, base, call) ||
		ctx.Flow().Tainted(base, flow.Path) || ctx.Flow().Tainted(semanticquery.CallArgument(call.Args, 0, "path"), flow.Path) {
		return 0, false, false
	}
	if sep, known := semanticquery.NativeString(ctx, concat.Right); known {
		return 0, false, sep == "/"
	}
	constant, ok := syntax.UnwrapParens(concat.Right).(*syntax.ConstFetch)
	return 0, false, ok && semanticquery.GlobalConstName(ctx, constant) == "DIRECTORY_SEPARATOR"
}
