// Package repeatedcookieheaderreplacement implements the native RepeatedCookieHeaderReplacement inspection.
package repeatedcookieheaderreplacement

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Append distinct cookie headers instead of replacing them."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "RepeatedCookieHeaderReplacement" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)
	var fixes []diagnostic.Fix

	if name != "header" {
		return
	}
	value, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "header"))
	if !known || !strings.HasPrefix(strings.ToLower(value), "set-cookie:") {
		return
	}
	replace := semanticquery.CallArgument(c.Args, 1, "replace")
	if replace != nil {
		truth, known := semanticquery.NativeTruth(ctx, replace)
		if !known || !truth {
			return
		}
	}
	previous := semanticquery.GlobalStateCall(ctx, c, "cookie-header")
	if previous == nil {
		return
	}
	old, _ := semanticquery.NativeString(ctx, semanticquery.CallArgument(previous.Args, 0, "header"))
	// Only two proven, different cookie names show a lost cookie: replacing
	// a cookie by a new value of itself is the intended use of the default.
	oldName, newName := cookieHeaderName(old), cookieHeaderName(value)
	if oldName == "" || newName == "" || oldName == newName || cookieHeaderRemovedBetween(ctx, previous, c) {
		return
	}
	if _, bound := ctx.Flow().BoundArguments(c); bound && replace == nil && semanticquery.CommentFreeSpan(ctx, c.Args.Span()) {
		insert := syntax.Span{Start: c.Args.Span().End - 1, End: c.Args.Span().End - 1}
		addition := ", false"
		for _, node := range c.Args.Args {
			if arg, ok := node.(*syntax.Arg); ok && arg.Name != nil {
				addition = ", replace: false"
			}
		}
		last := c.Args.Args[len(c.Args.Args)-1]
		tail := syntax.Span{Start: last.Span().End, End: c.Args.Span().End - 1}
		if strings.TrimSpace(string(ctx.File.Src[tail.Start:tail.End])) == "," {
			addition = strings.TrimPrefix(addition, ",")
		}
		fixes = append(fixes, semanticquery.ReplacementFix(insert, addition, "Append this cookie header"))
	}

	ctx.ReportNode(c, message, fixes...)
}

// cookieHeaderName returns the cookie name of a Set-Cookie field line, or ""
// when the line does not carry a name=value pair.
func cookieHeaderName(line string) string {
	_, pair, _ := strings.Cut(line, ":")
	pair, _, _ = strings.Cut(pair, ";")
	name, _, found := strings.Cut(pair, "=")
	if !found {
		return ""
	}
	return strings.TrimSpace(name)
}

// cookieHeaderRemovedBetween reports whether a header_remove() call that may
// drop the Set-Cookie headers sits between the two header() calls, on any
// branch.
func cookieHeaderRemovedBetween(ctx *analysis.Context, previous, c *syntax.FuncCall) bool {
	for _, reached := range ctx.Flow().Calls(syntax.EnclosingVariableScope(c)) {
		remove, ok := reached.Node.(*syntax.FuncCall)
		if !ok || semanticquery.NativeBuiltinName(ctx, remove) != "header_remove" {
			continue
		}
		if remove.Span().Start < previous.Span().End || remove.Span().Start >= c.Span().Start {
			continue
		}
		// Without an argument every header goes; an unknown name may be this one.
		target := semanticquery.CallArgument(remove.Args, 0, "name")
		field, known := semanticquery.NativeString(ctx, target)
		if target == nil || !known || strings.EqualFold(strings.TrimSpace(field), "set-cookie") {
			return true
		}
	}
	return false
}
