package nonsecureparsestrusage

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// nonSecureParseStrUsage reports one-argument parse_str()/mb_parse_str()
// calls, which create variables in the current scope.
type nonSecureParseStrUsage struct{}

func (nonSecureParseStrUsage) ID() string { return "NonSecureParseStrUsage" }
func (nonSecureParseStrUsage) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func (nonSecureParseStrUsage) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	name, ok := call.Name.(*syntax.Name)
	if !ok || call.Args == nil || len(call.Args.Args) != 1 {
		return
	}
	if fn := ctx.GlobalFunctionName(call); fn != "parse_str" && fn != "mb_parse_str" {
		return
	}
	ctx.Report(astquery.NamePartSpan(name), "Pass a result array as second argument instead of creating variables.")
}
