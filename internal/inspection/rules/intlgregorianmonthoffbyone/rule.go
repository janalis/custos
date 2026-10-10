// Package intlgregorianmonthoffbyone implements the native IntlGregorianMonthOffByOne inspection.
package intlgregorianmonthoffbyone

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use zero-based Gregorian month numbers."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "IntlGregorianMonthOffByOne" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall, syntax.KNew} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	var arg syntax.Expr
	switch c := n.(type) {
	case *syntax.MethodCall:
		if !semanticquery.NativeMethod(ctx, c, "IntlGregorianCalendar", "set") || c.Args == nil || len(c.Args.Args) < 3 {
			return
		}
		arg = semanticquery.CallArgument(c.Args, 1, "month")
	case *syntax.New:
		decl := semanticquery.NativeNewClass(ctx, c)
		if decl == nil || !strings.HasPrefix(decl.File, "stubs/") || !strings.EqualFold(decl.FQN, "IntlGregorianCalendar") || c.Args == nil || len(c.Args.Args) < 3 {
			return
		}
		arg = semanticquery.CallArgument(c.Args, 1, "month")
	}
	if _, ok := syntax.UnwrapParens(arg).(*syntax.Literal); !ok {
		return
	}
	month, ok := semanticquery.NativeInt(ctx, arg)
	if ok && month >= 1 && month <= 12 {
		ctx.ReportNode(n, message)
	}
}
