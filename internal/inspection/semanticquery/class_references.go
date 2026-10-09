package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// ReferencedClasses resolves a class reference (a class name, self/static/parent,
// or an expression whose inferred type holds classes) to class FQNs without
// leading backslash.
func ReferencedClasses(ctx *analysis.Context, ref syntax.Expr) []string {
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
