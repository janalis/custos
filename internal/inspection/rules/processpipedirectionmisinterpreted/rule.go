// Package processpipedirectionmisinterpreted implements the native ProcessPipeDirectionMisinterpreted inspection.
package processpipedirectionmisinterpreted

import (
	"custos/internal/inspection/analysis"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Choose pipe modes from the child perspective."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ProcessPipeDirectionMisinterpreted" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)
	if name != "fwrite" && name != "fread" && name != "stream_get_contents" {
		return
	}
	pipe, ok := syntax.UnwrapParens(semanticquery.CallArgument(c.Args, 0, "stream")).(*syntax.ArrayDimFetch)
	if !ok {
		return
	}
	if !semanticquery.ExpansionDUnaliased(ctx, pipe.Var, n) {
		return
	}
	open := semanticquery.ExpansionCOutputCall(ctx, c, pipe.Var, []string{"proc_open"}, 2, "pipes")
	if open == nil || !success(ctx, open, c) {
		return
	}
	key, known := semanticquery.NativeArrayKey(ctx, pipe.Dim)
	if !known {
		return
	}
	descriptors, known := semanticquery.NativeArrayEntries(ctx, semanticquery.CallArgument(open.Args, 1, "descriptor_spec"))
	if !known {
		return
	}
	tuple, known := semanticquery.NativeArrayEntries(ctx, descriptors[key])
	if !known {
		return
	}
	kind, kk := semanticquery.NativeString(ctx, tuple["i:0"])
	mode, mk := semanticquery.NativeString(ctx, tuple["i:1"])
	if kk && mk && kind == "pipe" && ((name == "fwrite" && mode == "w") || (name != "fwrite" && mode == "r")) {
		ctx.ReportNode(n, message)
	}
}

func success(ctx *analysis.Context, open *syntax.FuncCall, at syntax.Node) bool {
	for p := at; p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		b, ok := p.(*syntax.If)
		if !ok || !b.Body.Span().Contains(at.Span()) {
			continue
		}
		c, ok := syntax.UnwrapParens(b.Cond).(*syntax.FuncCall)
		if ok && semanticquery.NativeBuiltin(ctx, c, "is_resource") && semanticquery.NativeLocalValue(ctx, semanticquery.CallArgument(c.Args, 0, "value")) == open {
			return true
		}
	}
	return false
}
