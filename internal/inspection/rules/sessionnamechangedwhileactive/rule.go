// Package sessionnamechangedwhileactive implements the native SessionNameChangedWhileActive inspection.
package sessionnamechangedwhileactive

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Set the session name before starting the session."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SessionNameChangedWhileActive" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "session_name") || semanticquery.CallArgument(call.Args, 0, "name") == nil {
		return
	}
	state, known := ctx.Flow().GlobalStateBefore(call, "session")
	if !known || state.Operation != "session_start" {
		return
	}
	active := false
	for _, reached := range ctx.Flow().Calls(syntax.EnclosingVariableScope(call)) {
		start, ok := reached.Node.(*syntax.FuncCall)
		if !ok || start.Span() != state.Span {
			continue
		}
		active = true
		options := semanticquery.CallArgument(start.Args, 0, "options")
		if options != nil {
			values, known := semanticquery.NativeArrayEntries(ctx, options)
			if !known {
				return
			}
			if e, exists := values["s:read_and_close"]; exists {
				value, known := semanticquery.NativeTruth(ctx, e)
				if !known || value {
					return
				}
			}
		}
	}
	if active {
		ctx.ReportNode(call, message)
	}
}
