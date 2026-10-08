package probablebugs

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// simpleXMLLoadFileUsage reports simplexml_load_file() calls (PHP bug #62577)
// and rewrites them to simplexml_load_string(file_get_contents(...)).
type simpleXMLLoadFileUsage struct{}

func init() { register(simpleXMLLoadFileUsage{}) }

const simpleXMLLoadFileUsageMsg = "simplexml_load_file() is affected by PHP bug #62577; load the contents with file_get_contents() and parse them with simplexml_load_string()."

// Semantic marks the rule as needing the project index (user functions with
// the same names may be declared in other files).
func (simpleXMLLoadFileUsage) Semantic() {}

func (simpleXMLLoadFileUsage) ID() string { return "SimpleXmlLoadFileUsage" }

func (simpleXMLLoadFileUsage) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func (simpleXMLLoadFileUsage) Check(ctx *analysis.Context, n syntax.Node) {
	// custos: the bug needs the entity loader disabled with
	// libxml_disable_entity_loader(), deprecated from PHP 8.0 (libxml 2.9
	// disables external entities by default).
	if ctx.PHP >= phpver.PHP80 {
		return
	}
	call := n.(*syntax.FuncCall)
	if !strings.EqualFold(util.CallLastName(call), "simplexml_load_file") || call.Args == nil || len(call.Args.Args) == 0 { // D1, D2/E1
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
	ctx.ReportNode(call, simpleXMLLoadFileUsageMsg, analysis.Fix{
		Title: "Use simplexml_load_string(file_get_contents(...))",
		Edits: func() []analysis.TextEdit {
			var b strings.Builder
			b.WriteString(util.QualifiedBuiltin(ctx, "simplexml_load_string", call.Span().Start))
			b.WriteByte('(')
			b.WriteString(util.QualifiedBuiltin(ctx, "file_get_contents", call.Span().Start))
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
