package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// NativeDateClass proves a receiver is a builtin date object, excluding
// application subclasses whose methods can change the contract.
func NativeDateClass(ctx *analysis.Context, e syntax.Expr) string {
	classes := ctx.TypeOf(e).Classes()
	if len(classes) != 1 {
		return ""
	}
	class := strings.TrimPrefix(classes[0], `\`)
	if class != "DateTime" && class != "DateTimeImmutable" && class != "DateInterval" {
		return ""
	}
	if class == "DateTimeImmutable" && ctx.PHP < phpversion.PHP55 {
		return ""
	}
	if c := ctx.Index().Class(class, ctx.PHP); c == nil {
		return ""
	}
	return class
}

// NativeDateStatic recognizes a builtin date static method.
func NativeDateStatic(ctx *analysis.Context, n syntax.Node, method string) *syntax.StaticCall {
	call, ok := n.(*syntax.StaticCall)
	if !ok {
		return nil
	}
	name, ok := call.Name.(*syntax.Identifier)
	if !ok || !strings.EqualFold(name.Value, method) {
		return nil
	}
	class := StaticCallClass(ctx, call.Class)
	if class != "DateTime" && class != "DateTimeImmutable" {
		return nil
	}
	if class == "DateTimeImmutable" && ctx.PHP < phpversion.PHP55 {
		return nil
	}
	if c := ctx.Index().Class(class, ctx.PHP); c == nil {
		return nil
	}
	return call
}

// NativeDateFormat locates a date-format argument in builtin formatting calls.
func NativeDateFormat(ctx *analysis.Context, n syntax.Node) syntax.Expr {
	if c, _ := GlobalCall(ctx, n, "date", "gmdate"); c != nil {
		return CallArgument(c.Args, 0, "format")
	}
	c, ok := n.(*syntax.MethodCall)
	if !ok {
		return nil
	}
	name, ok := c.Name.(*syntax.Identifier)
	if !ok || !strings.EqualFold(name.Value, "format") || NativeDateClass(ctx, c.Var) == "" || NativeDateClass(ctx, c.Var) == "DateInterval" {
		return nil
	}
	return CallArgument(c.Args, 0, "format")
}

// NativeFormatTokens marks unescaped date-format bytes.
func NativeFormatTokens(format string) map[byte]bool {
	tokens := map[byte]bool{}
	for i := 0; i < len(format); i++ {
		if format[i] == '\\' {
			i++
			continue
		}
		tokens[format[i]] = true
	}
	return tokens
}
