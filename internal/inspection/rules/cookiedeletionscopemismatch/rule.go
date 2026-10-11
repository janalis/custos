// Package cookiedeletionscopemismatch implements the native CookieDeletionScopeMismatch inspection.
package cookiedeletionscopemismatch

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Delete the cookie using its original path and domain."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CookieDeletionScopeMismatch" }
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
	value, vk := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 1, "value"))
	expiry, pathArg, domainArg, ok := settings(ctx, call)
	expires, knownExpiry := semanticquery.NativeInt(ctx, expiry)
	if !known || !vk || value != "" || !ok || !knownExpiry || expires <= 0 || expires > 1 {
		return
	}
	stmt, ok := call.Parent().(*syntax.ExprStmt)
	if !ok {
		return
	}
	prev, ok := astquery.PrevStmtNoDoc(ctx.File, stmt)
	if !ok {
		return
	}
	st, ok := prev.(*syntax.ExprStmt)
	if !ok {
		return
	}
	first, ok := st.Expr.(*syntax.FuncCall)
	if !ok {
		return
	}
	firstName := semanticquery.NativeBuiltinName(ctx, first)
	if firstName != "setcookie" && firstName != "setrawcookie" {
		return
	}
	stored, sk := semanticquery.NativeString(ctx, semanticquery.CallArgument(first.Args, 0, "name"))
	content, ck := semanticquery.NativeString(ctx, semanticquery.CallArgument(first.Args, 1, "value"))
	if !sk || !ck || content == "" || stored != cookie {
		return
	}
	_, oldPathArg, oldDomainArg, ok := settings(ctx, first)
	if !ok {
		return
	}
	path, pk := scopePart(ctx, pathArg)
	domain, dk := scopePart(ctx, domainArg)
	oldPath, opk := scopePart(ctx, oldPathArg)
	oldDomain, odk := scopePart(ctx, oldDomainArg)
	if !pk || !dk || !opk || !odk {
		return
	}
	if differs(path, oldPath) || differs(cookieDomain(domain), cookieDomain(oldDomain)) {
		ctx.ReportNode(call, message)
	}
}

// settings returns the expiry, path and domain expressions of a cookie setter,
// read from the options array or from the positional/named arguments. A nil
// expression is an omitted setting; ok is false when the form is not proven.
func settings(ctx *analysis.Context, c *syntax.FuncCall) (expires, path, domain syntax.Expr, ok bool) {
	value := semanticquery.CallArgument(c.Args, 2, "expires_or_options")
	if semanticquery.NativeArray(ctx, value) == nil {
		path = semanticquery.CallArgument(c.Args, 3, "path")
		domain = semanticquery.CallArgument(c.Args, 4, "domain")
		// PHP rejects any further argument next to an options array, so one
		// of them proves the scalar form even when the expiry is computed.
		scalar := value == nil || path != nil || domain != nil ||
			semanticquery.CallArgument(c.Args, 5, "secure") != nil ||
			semanticquery.CallArgument(c.Args, 6, "httponly") != nil
		if !scalar {
			_, scalar = semanticquery.NativeInt(ctx, value)
		}
		return value, path, domain, scalar
	}
	entries, ok := semanticquery.NativeArrayEntries(ctx, value)
	if !ok {
		return nil, nil, nil, false
	}
	// Option names are matched case-insensitively by PHP; two spellings of
	// one name leave the effective value to evaluation order, so give up.
	options := make(map[string]syntax.Expr, len(entries))
	for key, e := range entries {
		name := strings.ToLower(key)
		if _, dup := options[name]; dup {
			return nil, nil, nil, false
		}
		options[name] = e
	}
	return options["s:expires"], options["s:path"], options["s:domain"], true
}

// scopePart returns the literal text of a path or domain setting; an omitted
// setting is the empty string, which PHP also treats as "not sent".
func scopePart(ctx *analysis.Context, e syntax.Expr) (string, bool) {
	if e == nil {
		return "", true
	}
	return semanticquery.NativeString(ctx, e)
}

// cookieDomain folds the spellings browsers treat as one domain.
func cookieDomain(domain string) string {
	return strings.TrimPrefix(strings.ToLower(domain), ".")
}

// differs reports two explicit, different values. An omitted value falls back
// to a request-dependent default and is never a proven mismatch.
func differs(a, b string) bool {
	return a != "" && b != "" && a != b
}
