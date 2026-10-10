// Package filereadfailureunchecked implements the native FileReadFailureUnchecked inspection.
package filereadfailureunchecked

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Handle file reading failure before using its result."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FileReadFailureUnchecked" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall, syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	var inputs []syntax.Expr
	switch x := n.(type) {
	case *syntax.FuncCall:
		inputs = semanticquery.NativeStringInputs(ctx, x, x.Args)
	case *syntax.Binary:
		if x.Op.Kind != syntax.TDot {
			return
		}
		inputs = []syntax.Expr{x.Left, x.Right}
	}
	for _, input := range inputs {
		inner, ok := semanticquery.NativeValue(ctx, input).(*syntax.FuncCall)
		if !ok {
			continue
		}
		switch semanticquery.NativeBuiltinName(ctx, inner) {
		case "file_get_contents", "stream_get_contents":
		default:
			continue
		}
		if ctx.TypeOf(input).OnlyOf("string") {
			continue
		}
		ctx.ReportNode(n, message)
		return
	}
}
