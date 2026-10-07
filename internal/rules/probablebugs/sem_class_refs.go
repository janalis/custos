package probablebugs

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/syntax"
)

// semClassesOf resolves a class reference (a class name, self/static/parent,
// or an expression whose inferred type holds classes) to class FQNs without
// leading backslash.
func semClassesOf(ctx *analysis.Context, ref syntax.Expr) []string {
	if n, ok := ref.(*syntax.Name); ok {
		var fqn string
		switch strings.ToLower(n.Value) {
		case "self", "static":
			fqn = ctx.Names().DeclFQN(syntax.EnclosingClass(n))
		case "parent":
			fqn = ctx.Names().ParentFQN(syntax.EnclosingClass(n))
		default:
			fqn = ctx.Names().Class(n.Value, n.Span().Start)
		}
		if fqn == "" {
			return nil
		}
		return []string{fqn}
	}
	var out []string
	for _, c := range ctx.TypeOf(ref).Classes() {
		out = append(out, strings.TrimPrefix(c, `\`))
	}
	return out
}

// semArgs returns the argument nodes of a call's argument list (nil list ok).
func semArgs(list *syntax.ArgList) []syntax.Expr {
	if list == nil {
		return nil
	}
	return list.Args
}

// semArgValue returns the value of a plain argument entry.
func semArgValue(a syntax.Expr) syntax.Expr {
	if arg, ok := a.(*syntax.Arg); ok {
		return arg.Value
	}
	return nil
}
