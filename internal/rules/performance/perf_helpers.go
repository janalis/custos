package performance

import (
	"custos/internal/analysis"
	"custos/internal/syntax"
)

// perfCall returns e as a plain function call that resolves to one of the
// global functions names (lower-case; compared case-insensitively, as PHP
// does, and only when the call reaches the global function rather than a
// same-named namespaced or imported one), plus that lower-case name.
func perfCall(ctx *analysis.Context, e syntax.Node, names ...string) (*syntax.FuncCall, string) {
	call, ok := e.(*syntax.FuncCall)
	if !ok {
		return nil, ""
	}
	g := ctx.GlobalFunctionName(call)
	if g == "" {
		return nil, ""
	}
	for _, n := range names {
		if g == n {
			return call, n
		}
	}
	return nil, ""
}

// perfArgs returns the call's arguments as written (spreads and named
// arguments included); nil when a placeholder `f(...)` or no list is present.
func perfArgs(list *syntax.ArgList) []*syntax.Arg {
	if list == nil {
		return nil
	}
	out := make([]*syntax.Arg, 0, len(list.Args))
	for _, a := range list.Args {
		arg, ok := a.(*syntax.Arg)
		if !ok || arg.Value == nil {
			return nil
		}
		out = append(out, arg)
	}
	return out
}

// perfPlainArgs returns the argument values when none is spread or named.
func perfPlainArgs(list *syntax.ArgList) ([]syntax.Expr, bool) {
	args := perfArgs(list)
	if list == nil || len(args) != len(list.Args) {
		return nil, false
	}
	out := make([]syntax.Expr, len(args))
	for i, a := range args {
		if a.Unpack || a.Name != nil {
			return nil, false
		}
		out[i] = a.Value
	}
	return out, true
}

// perfPlainVar returns the name of a simple `$name` variable.
func perfPlainVar(e syntax.Node) (string, bool) {
	v, ok := e.(*syntax.Variable)
	if !ok || v.NameExpr != nil || v.Name == "" {
		return "", false
	}
	return v.Name, true
}

// perfIsFalse reports whether e is the unqualified constant false (any case).
func perfIsFalse(e syntax.Node) bool {
	c, ok := e.(*syntax.ConstFetch)
	if !ok || c.Name == nil || c.Name.NameKind != syntax.NameUnqualified {
		return false
	}
	v := c.Name.Value
	return len(v) == 5 && (v[0]|0x20) == 'f' && (v[1]|0x20) == 'a' && (v[2]|0x20) == 'l' && (v[3]|0x20) == 's' && (v[4]|0x20) == 'e'
}
