// Package cookiedeletionscopemismatch implements the native CookieDeletionScopeMismatch inspection.
package cookiedeletionscopemismatch

import (
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
	expiry, ek := setting(ctx, call, 2, "expires")
	expires, knownExpiry := semanticquery.NativeInt(ctx, expiry)
	if !known || !vk || value != "" || !ek || !knownExpiry || expires <= 0 || expires > 1 {
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
	path, pk := stringSetting(ctx, call, 3, "path")
	domain, dk := stringSetting(ctx, call, 4, "domain")
	oldPath, opk := stringSetting(ctx, first, 3, "path")
	oldDomain, odk := stringSetting(ctx, first, 4, "domain")
	if pk && dk && opk && odk && (path != oldPath || domain != oldDomain) {
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
		return entries["s:"+name], true
	}
	if pos == 2 {
		return value, true
	}
	if _, ok := semanticquery.NativeInt(ctx, value); ok {
		return semanticquery.CallArgument(c.Args, pos, name), true
	}
	return nil, false
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
