// Package unknownnamedargument implements the UnknownNamedArgument inspection.
package unknownnamedargument

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

type rule struct{}

func New() analysis.Rule { return rule{} }
func (rule) ID() string  { return "UnknownNamedArgument" }
func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall, syntax.KMethodCall, syntax.KStaticCall}
}

const message = "Use a declared parameter name."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP80 {
		return
	}
	target, known := semanticquery.ResolveCallee(ctx, n)
	if !known {
		return
	}
	names := map[string]bool{}
	for _, param := range target.Params {
		if param.Variadic {
			return
		}
		names[param.Name] = true
	}
	var args *syntax.ArgList
	switch call := n.(type) {
	case *syntax.FuncCall:
		args = call.Args
	case *syntax.MethodCall:
		args = call.Args
	case *syntax.StaticCall:
		args = call.Args
	}
	if args == nil {
		return
	}
	for _, node := range args.Args {
		if arg, ok := node.(*syntax.Arg); ok && arg.Name != nil && !names[arg.Name.Value] {
			ctx.ReportNode(arg, message)
		}
	}
}

func (rule) Semantic() {}
