package util

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/names"
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

// StringCallableClass spells the class named in a string callable (`'Foo'`
// in `['Foo', 'm']` or `'Foo::m'`) so that `Class::m()` written at offset at
// names the same class: a leading `\` is added when the name, written as-is,
// would resolve to another class. self/static/parent are kept.
func StringCallableClass(ctx *analysis.Context, class string, at uint32) string {
	if strings.HasPrefix(class, `\`) || class == "" || names.IsSpecialClass(class) {
		return class
	}
	if strings.EqualFold(ctx.Names().Class(class, at), class) {
		return class
	}
	return `\` + class
}
