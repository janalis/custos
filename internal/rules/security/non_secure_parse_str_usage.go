package security

import (
	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// nonSecureParseStrUsage reports one-argument parse_str()/mb_parse_str()
// calls, which create variables in the current scope.
type nonSecureParseStrUsage struct{}

func init() { register(nonSecureParseStrUsage{}) }

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
	ctx.Report(namePartSpan(name), "Pass a result array as second argument instead of creating variables.")
}

// namePartSpan returns the span of the last segment of a written name.
func namePartSpan(name *syntax.Name) syntax.Span {
	s := name.Span()
	part := util.LastNamePart(name.Value)
	return syntax.Span{Start: s.End - uint32(len(part)), End: s.End}
}
