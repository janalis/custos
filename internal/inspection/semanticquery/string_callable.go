package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
)

// StringCallableFunction spells the function named by a string callable fn
// (decoded contents) so that a direct call written at offset at reaches the
// same function. A string callable always holds an absolute name, while a
// written call resolves against the namespace and imports: the name gets a
// leading `\` when, written as-is, it would resolve elsewhere (a relative
// qualified name, an import, or a same-named function in the namespace).
func StringCallableFunction(ctx *analysis.Context, fn string, at uint32) string {
	if strings.HasPrefix(fn, `\`) || fn == "" {
		return fn
	}
	if !strings.Contains(fn, `\`) {
		return QualifiedBuiltin(ctx, fn, at)
	}
	if fqn, _ := ctx.Names().Function(fn, at); strings.EqualFold(fqn, fn) {
		return fn
	}
	return `\` + fn
}
