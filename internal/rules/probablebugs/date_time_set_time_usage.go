package probablebugs

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// dateTimeSetTimeUsage reports the microseconds argument of
// DateTime::setTime()/date_time_set() when targeting PHP < 7.1.
type dateTimeSetTimeUsage struct{}

func init() { register(dateTimeSetTimeUsage{}) }

func (dateTimeSetTimeUsage) ID() string { return "DateTimeSetTimeUsage" }

func (dateTimeSetTimeUsage) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KMethodCall, syntax.KStaticCall, syntax.KFuncCall}
}

const dateTimeSetTimeMsg = "Microseconds argument requires PHP 7.1+; on this version the call returns false."

func (dateTimeSetTimeUsage) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP >= phpver.PHP71 { // E1
		return
	}
	var name syntax.Expr
	var args []syntax.Expr
	var classes func() []string
	switch x := n.(type) {
	case *syntax.MethodCall:
		name, args = x.Name, semArgs(x.Args)
		classes = func() []string { return semClassesOf(ctx, x.Var) }
	case *syntax.StaticCall:
		name, args = x.Name, semArgs(x.Args)
		classes = func() []string { return semClassesOf(ctx, x.Class) }
	case *syntax.FuncCall: // D4-D6
		if !strings.EqualFold(util.CallLastName(x), "date_time_set") || util.ArgCount(x) != 5 {
			return
		}
		if !util.ResolvesToGlobalFunction(ctx.Names(), ctx.Index(), ctx.PHP, x, "date_time_set") {
			return
		}
		ctx.ReportNode(x.Args.Args[4], dateTimeSetTimeMsg)
		return
	}
	// D1-D3
	id, ok := name.(*syntax.Identifier)
	if !ok || !strings.EqualFold(id.Value, "setTime") || len(args) != 4 {
		return
	}
	for _, cls := range classes() {
		if m := ctx.Index().FindMethod(cls, "setTime", ctx.PHP); m != nil && strings.EqualFold(strings.TrimPrefix(m.Class, `\`), "DateTime") {
			ctx.ReportNode(args[3], dateTimeSetTimeMsg)
			return
		}
	}
}

// Semantic marks the rule as needing the project index.
func (dateTimeSetTimeUsage) Semantic() {}
