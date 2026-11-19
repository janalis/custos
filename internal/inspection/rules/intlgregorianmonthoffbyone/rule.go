// Package intlgregorianmonthoffbyone implements the native IntlGregorianMonthOffByOne inspection.
package intlgregorianmonthoffbyone

import (
	"strings"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/flowquery"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
)

const message = "Use zero-based Gregorian month numbers."

type rule struct{}

// New returns the IntlGregorianMonthOffByOne inspection.
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
