package probablebugs

import (
	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// pregQuoteUsage reports preg_quote() calls without the delimiter argument.
type pregQuoteUsage struct{}

func init() { register(pregQuoteUsage{}) }

// Semantic marks the rule as needing the project index (a user function
// named preg_quote may be declared in another file).
func (pregQuoteUsage) Semantic() {}

func (pregQuoteUsage) ID() string { return "PregQuoteUsage" }

func (pregQuoteUsage) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func (pregQuoteUsage) Check(ctx *analysis.Context, n syntax.Node) {
	call, ok := util.IsFuncNamedFold(n, "preg_quote")        // D1
	if !ok || call.Args == nil || len(call.Args.Args) != 1 { // D2
		return
	}
	if _, ok := call.Args.Args[0].(*syntax.Arg); !ok {
		return // first-class callable syntax preg_quote(...)
	}
	if !util.ResolvesToGlobalFunction(ctx.Names(), ctx.Index(), ctx.PHP, call, "preg_quote") { // D1
		return
	}
	ctx.Report(util.NamePartSpan(call.Name.(*syntax.Name)), "Pass the pattern delimiter to preg_quote() as its second argument.")
}
