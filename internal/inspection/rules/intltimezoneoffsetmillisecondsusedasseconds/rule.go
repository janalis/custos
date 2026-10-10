// Package intltimezoneoffsetmillisecondsusedasseconds implements the native IntlTimezoneOffsetMillisecondsUsedAsSeconds inspection.
package intltimezoneoffsetmillisecondsusedasseconds

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Convert timezone offsets to seconds."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "IntlTimezoneOffsetMillisecondsUsedAsSeconds" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "DateTime", "modify") && !semanticquery.NativeMethod(ctx, c, "DateTimeImmutable", "modify") {
		return
	}
	arg, ok := syntax.UnwrapParens(semanticquery.CallArgument(c.Args, 0, "modifier")).(*syntax.Binary)
	if !ok || arg.Op.Kind != syntax.TDot {
		return
	}
	suffix, known := semanticquery.NativeString(ctx, arg.Right)
	if !known || strings.TrimSpace(suffix) != "seconds" {
		return
	}
	offset := syntax.UnwrapParens(arg.Left)
	var vars []syntax.Expr
	if sum, ok := offset.(*syntax.Binary); ok {
		if sum.Op.Kind != syntax.TPlus {
			return
		}
		vars = []syntax.Expr{sum.Left, sum.Right}
	} else {
		vars = []syntax.Expr{offset}
	}
	found := false
	for _, fact := range ctx.Flow().Calls(syntax.EnclosingVariableScope(c)) {
		source, ok := fact.Node.(*syntax.MethodCall)
		if !ok || !semanticquery.NativeMethod(ctx, source, "IntlTimeZone", "getOffset") || !semanticquery.NativeDominates(source, c) {
			continue
		}
		all := true
		for _, v := range vars {
			if !offsetUnchanged(ctx, source, arg, v) || !semanticquery.ExpansionCSame(ctx, v, semanticquery.CallArgument(source.Args, 2, "rawOffset")) && !semanticquery.ExpansionCSame(ctx, v, semanticquery.CallArgument(source.Args, 3, "dstOffset")) {
				all = false
			}
		}
		found = found || all
	}
	if found {
		ctx.ReportNode(c, message, astquery.ReplaceFix(arg.Left.Span(), "("+ctx.Text(arg.Left)+" / 1000)"))
	}
}

// An out parameter loses its unit proof after any intervening use that may
// change it. Reject aliases conservatively, including ones created earlier.
func offsetUnchanged(ctx *analysis.Context, source *syntax.MethodCall, at syntax.Node, value syntax.Expr) bool {
	variable, ok := syntax.UnwrapParens(value).(*syntax.Variable)
	if !ok {
		return false
	}
	valid, budget := true, 256
	scope := syntax.EnclosingVariableScope(at)
	visit := func(n syntax.Node) bool {
		budget--
		if budget < 0 {
			valid = false
			return false
		}
		if n.Span().Start >= at.Span().Start {
			return false
		}
		if n != scope && syntax.IsVariableScope(n) {
			valid = false
			return false
		}
		if a, ok := n.(*syntax.Assign); ok && a.ByRef {
			valid = false
			return false
		}
		if v, ok := n.(*syntax.Variable); ok && v.Name == variable.Name && v.Span().Start > source.Span().End {
			valid = false
		}
		return true
	}
	if scope != nil {
		syntax.Inspect(scope, visit)
	} else {
		syntax.InspectFile(ctx.File, visit)
	}
	return valid
}
