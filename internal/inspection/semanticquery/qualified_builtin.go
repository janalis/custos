package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// QualifiedBuiltin returns how a fix should spell a call to the global
// function name (lower-case) inserted at offset at: the bare name when an
// unqualified call there reaches the global function, `\name` when it would
// not — a `use function` import of another function under that name, or a
// same-named function declared in the current namespace (this file or the
// project index), which PHP prefers over the global fallback.
func QualifiedBuiltin(ctx *analysis.Context, name string, at uint32) string {
	if BareReachesGlobal(ctx, name, at) {
		return name
	}
	return `\` + name
}

// BareReachesGlobal reports whether an unqualified call to the global
// function name written at offset at resolves to that global function.
func BareReachesGlobal(ctx *analysis.Context, name string, at uint32) bool {
	fqn, fallback := ctx.Names().Function(name, at)
	if fallback != "" {
		return ctx.Index().Function(fqn, ctx.PHP) == nil
	}
	return strings.EqualFold(fqn, name)
}

// QualifiedBuiltinFor is QualifiedBuiltin for a fix rewriting call: the
// emitted name also keeps a leading backslash when call was written fully
// qualified.
func QualifiedBuiltinFor(ctx *analysis.Context, name string, call *syntax.FuncCall) string {
	if n, ok := call.Name.(*syntax.Name); ok && strings.HasPrefix(n.Value, `\`) {
		return `\` + name
	}
	return QualifiedBuiltin(ctx, name, call.Span().Start)
}
