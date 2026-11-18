package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/stubs"
)

// IsBuiltinClass reports whether c is the declaration from the embedded PHP
// stubs (as opposed to a user/project declaration).
func IsBuiltinClass(c *index.Class, ver phpversion.Version) bool {
	if c == nil {
		return false
	}
	return stubs.Index().Class(c.FQN, ver) == c
}

// BuiltinClassConstant matches a resolved builtin class constant, including aliases.
func BuiltinClassConstant(ctx *analysis.Context, e syntax.Expr, class, name string) bool {
	c, ok := syntax.UnwrapParens(e).(*syntax.ClassConstFetch)
	if !ok {
		return false
	}
	id, ok := c.Name.(*syntax.Identifier)
	if !ok || id.Value != name || !strings.EqualFold(StaticCallClass(ctx, c.Class), class) {
		return false
	}
	decl := ctx.Index().Class(class, ctx.PHP)
	return decl != nil && strings.HasPrefix(decl.File, "stubs/")
}

// BuiltinStaticCall recognizes a builtin static method without treating lookalikes as APIs.
func BuiltinStaticCall(ctx *analysis.Context, e syntax.Expr, class, method string) bool {
	c, ok := e.(*syntax.StaticCall)
	if !ok || !strings.EqualFold(StaticCallClass(ctx, c.Class), class) {
		return false
	}
	id, ok := c.Name.(*syntax.Identifier)
	if !ok || !strings.EqualFold(id.Value, method) {
		return false
	}
	m := ctx.Index().FindMethod(class, method, ctx.PHP)
	return m != nil && m.Builtin && m.Avail.In(ctx.PHP)
}

// IsBuiltinConst reports whether e reads the builtin constant name.
func IsBuiltinConst(ctx *analysis.Context, e syntax.Expr, name string) bool {
	c, ok := syntax.UnwrapParens(e).(*syntax.ConstFetch)
	if !ok {
		return false
	}
	resolved := GlobalConstName(ctx, c)
	symbol := ctx.Index().Constant(resolved, ctx.PHP)
	return resolved == name && symbol != nil && symbol.Builtin
}

// ReadsGlobalConstant recognizes an unshadowed named extension constant.
func ReadsGlobalConstant(ctx *analysis.Context, e syntax.Expr, name string) bool {
	c, ok := e.(*syntax.ConstFetch)
	return ok && GlobalConstName(ctx, c) == name
}

// NamesGlobalConstant reports whether e evaluates to PHP's global constant
// name, written bare where it reaches the global one.
func NamesGlobalConstant(ctx *analysis.Context, e syntax.Expr, name string) bool {
	c, ok := NativeValue(ctx, e).(*syntax.ConstFetch)
	return ok && c.Name != nil && strings.EqualFold(c.Name.Value, name) && BareReachesGlobal(ctx, name, c.Span().Start)
}
