// Package samesitenonewithoutsecure implements the native SameSiteNoneWithoutSecure inspection.
package samesitenonewithoutsecure

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Use a Secure cookie with SameSite None."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SameSiteNoneWithoutSecure" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)
	if name != "setcookie" || ctx.PHP < phpversion.PHP73 {
		return
	}
	entries, known := semanticquery.NativeArrayEntries(ctx, semanticquery.CallArgument(c.Args, 2, "expires_or_options"))
	if !known {
		return
	}
	// PHP matches cookie option names without regard to case. Two spellings of
	// one option leave the effective value to entry order, so nothing is proven.
	options := make(map[string]syntax.Expr, len(entries))
	for key, value := range entries {
		key = strings.ToLower(key)
		if _, dup := options[key]; dup {
			return
		}
		options[key] = value
	}
	site, s := semanticquery.NativeString(ctx, options["s:samesite"])
	if !s || !strings.EqualFold(site, "none") {
		return
	}
	if secure, exists := options["s:secure"]; exists {
		if truth, known := semanticquery.NativeTruth(ctx, secure); !known || truth {
			return
		}
	}
	ctx.ReportNode(c, message)
}
