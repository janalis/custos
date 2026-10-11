// Package hostcookieprefixcontractviolation implements the native HostCookiePrefixContractViolation inspection.
package hostcookieprefixcontractviolation

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Set Secure and Path=/ and omit Domain for __Host- cookies."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "HostCookiePrefixContractViolation" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, call)
	if name != "setcookie" && name != "setrawcookie" {
		return
	}
	cookie, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 0, "name"))
	if !known || !strings.HasPrefix(cookie, "__Host-") {
		return
	}
	secure, sk := boolSetting(ctx, call, 5, "secure")
	path, pk := stringSetting(ctx, call, 3, "path")
	domain, dk := stringSetting(ctx, call, 4, "domain")
	if sk && pk && dk && (!secure || path != "/" || domain != "") {
		ctx.ReportNode(call, message)
	}
}

func setting(ctx *analysis.Context, c *syntax.FuncCall, pos int, name string) (syntax.Expr, bool) {
	value := semanticquery.CallArgument(c.Args, 2, "expires_or_options")
	if value == nil {
		return semanticquery.CallArgument(c.Args, pos, name), true
	}
	if a := semanticquery.NativeArray(ctx, value); a != nil {
		entries, ok := semanticquery.NativeArrayEntries(ctx, value)
		if !ok {
			return nil, false
		}
		return option(entries, name)
	}
	if _, ok := semanticquery.NativeInt(ctx, value); ok {
		return semanticquery.CallArgument(c.Args, pos, name), true
	}
	return nil, false
}

// option reads one options-array entry the way setcookie() does: key names
// are matched without regard to ASCII case. Several spellings of one name
// leave the effective value unproven.
func option(entries map[string]syntax.Expr, name string) (syntax.Expr, bool) {
	var found syntax.Expr
	seen := false
	for key, value := range entries {
		if !asciiEqualFold(key, "s:"+name) {
			continue
		}
		if seen {
			return nil, false
		}
		found, seen = value, true
	}
	return found, true
}

// asciiEqualFold compares against an already lower-case ASCII name.
func asciiEqualFold(s, lower string) bool {
	if len(s) != len(lower) {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		if c != lower[i] {
			return false
		}
	}
	return true
}

func stringSetting(ctx *analysis.Context, c *syntax.FuncCall, pos int, name string) (string, bool) {
	e, ok := setting(ctx, c, pos, name)
	if !ok {
		return "", false
	}
	if e == nil {
		return "", true
	}
	return semanticquery.NativeString(ctx, e)
}

func boolSetting(ctx *analysis.Context, c *syntax.FuncCall, pos int, name string) (bool, bool) {
	e, ok := setting(ctx, c, pos, name)
	if !ok {
		return false, false
	}
	if e == nil {
		return false, true
	}
	return semanticquery.NativeTruth(ctx, e)
}
