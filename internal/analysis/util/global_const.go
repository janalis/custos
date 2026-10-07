package util

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/syntax"
)

// GlobalConstName returns the name of the global constant a constant
// reference resolves to (no leading backslash, as written), or "" when it
// targets a namespaced constant: a qualified non-global name
// (`Foo\PHP_VERSION`), a `use const` import of a namespaced constant, or an
// unqualified name in a namespace declaring a constant of that name (it
// wins over PHP's global fallback).
func GlobalConstName(ctx *analysis.Context, c *syntax.ConstFetch) string {
	if c == nil || c.Name == nil {
		return ""
	}
	fqn, fallback := ctx.Names().Const(c.Name.Value, c.Name.Span().Start)
	if fallback != "" {
		if ctx.Index().Constant(fqn, ctx.PHP) != nil {
			return ""
		}
		return fallback
	}
	if strings.Contains(fqn, `\`) {
		return ""
	}
	return fqn
}

// QualifiedGlobalConst returns how a fix should spell the global constant
// name inserted at offset at: the bare name when an unqualified reference
// there reaches the global constant, `\name` when a `use const` import or a
// constant declared in the current namespace would capture it.
func QualifiedGlobalConst(ctx *analysis.Context, name string, at uint32) string {
	fqn, fallback := ctx.Names().Const(name, at)
	if fallback != "" {
		if ctx.Index().Constant(fqn, ctx.PHP) == nil {
			return name
		}
	} else if fqn == name {
		return name
	}
	return `\` + name
}
