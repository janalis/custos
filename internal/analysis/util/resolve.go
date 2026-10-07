package util

import (
	"strings"

	"custos/internal/index"
	"custos/internal/names"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// ResolvedFunctionFQN resolves the named function call with PHP's runtime
// rules (imports, namespaced candidate first, then the global fallback)
// against ix and returns the FQN of the target as declared (no leading
// backslash). ok is false for dynamic calls and unresolvable names.
func ResolvedFunctionFQN(r *names.Resolver, ix *index.Index, ver phpver.Version, call *syntax.FuncCall) (string, bool) {
	name, ok := call.Name.(*syntax.Name)
	if !ok {
		return "", false
	}
	fqn, fallback := r.Function(name.Value, name.Span().Start)
	f := ix.ResolveFunction(fqn, fallback, ver)
	if f == nil {
		return "", false
	}
	return strings.TrimPrefix(f.FQN, `\`), true
}

// ResolvesToGlobalFunction reports whether call targets the global function
// global (case-insensitive) once namespaces, imports and the functions known
// to ix are taken into account: a same-named function in the current
// namespace or a non-global import wins over the global one.
func ResolvesToGlobalFunction(r *names.Resolver, ix *index.Index, ver phpver.Version, call *syntax.FuncCall, global string) bool {
	fqn, ok := ResolvedFunctionFQN(r, ix, ver, call)
	return ok && strings.EqualFold(fqn, global)
}
