package probablebugs

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// simpleXmlLoadFileUsage reports simplexml_load_file() calls (PHP bug #62577)
// and rewrites them to simplexml_load_string(file_get_contents(...)).
type simpleXmlLoadFileUsage struct{}

func init() { register(simpleXmlLoadFileUsage{}) }

const simpleXmlLoadFileUsageMsg = "simplexml_load_file() is affected by PHP bug #62577; load the contents with file_get_contents() and parse them with simplexml_load_string()."

// Semantic marks the rule as needing the project index (user functions with
// the same names may be declared in other files).
func (simpleXmlLoadFileUsage) Semantic() {}

func (simpleXmlLoadFileUsage) ID() string { return "SimpleXmlLoadFileUsage" }

func (simpleXmlLoadFileUsage) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func (simpleXmlLoadFileUsage) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !strings.EqualFold(writtenCallName(call), "simplexml_load_file") || call.Args == nil || len(call.Args.Args) == 0 { // D1, D2/E1
		return
	}
	if !util.ResolvesToGlobalFunction(ctx.Names(), ctx.Index(), ctx.PHP, call, "simplexml_load_file") { // D1
		return
	}
	// Divergence: skip spreads, named arguments and `f(...)`.
	args := make([]*syntax.Arg, 0, len(call.Args.Args))
	for _, a := range call.Args.Args {
		arg, ok := a.(*syntax.Arg)
		if !ok || arg.Unpack || arg.Name != nil {
			return
		}
		args = append(args, arg)
	}
	ctx.ReportNode(call, simpleXmlLoadFileUsageMsg, analysis.Fix{
		Title: "Use simplexml_load_string(file_get_contents(...))",
		Edits: func() []analysis.TextEdit {
			var b strings.Builder
			b.WriteString(simpleXmlGlobalName(ctx, call, "simplexml_load_string"))
			b.WriteByte('(')
			b.WriteString(simpleXmlGlobalName(ctx, call, "file_get_contents"))
			b.WriteByte('(')
			b.WriteString(ctx.Text(args[0]))
			b.WriteByte(')')
			for _, a := range args[1:] {
				b.WriteString(", ")
				b.WriteString(ctx.Text(a))
			}
			b.WriteByte(')')
			return []analysis.TextEdit{{Span: call.Span(), NewText: b.String()}}
		},
	})
}

// writtenCallName returns the last segment of a plain function call's name as
// written (case preserved), or "" for dynamic calls.
func writtenCallName(call *syntax.FuncCall) string {
	name, ok := call.Name.(*syntax.Name)
	if !ok {
		return ""
	}
	return name.Value[strings.LastIndexByte(name.Value, '\\')+1:]
}

// simpleXmlGlobalName returns fn unqualified when a bare call at the position
// of call reaches the global function, `\fn` otherwise (F1).
func simpleXmlGlobalName(ctx *analysis.Context, call *syntax.FuncCall, fn string) string {
	return util.QualifiedBuiltin(ctx, fn, call.Span().Start)
}
