// Package sessionlockheldduringblockingcall implements the native SessionLockHeldDuringBlockingCall inspection.
package sessionlockheldduringblockingcall

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Release the session lock before blocking when safe."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SessionLockHeldDuringBlockingCall" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	check(ctx, n, message)
}

func check(ctx *analysis.Context, n syntax.Node, message string) {
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)
	if name != "sleep" && name != "usleep" && name != "curl_exec" {
		return
	}
	start := semanticquery.GlobalStateCall(ctx, c, "session")
	if start == nil || !semanticquery.NativeBuiltin(ctx, start, "session_start") {
		return
	}
	if sessionStartKeepsLock(ctx, start) && sessionFilesHandler(ctx, start) {
		ctx.ReportNode(c, message)
	}
}

// sessionStartKeepsLock reports whether the session_start call provably leaves
// the session open: its options, when given, must be a literal array with
// constant string keys whose read_and_close entry is absent or known falsy.
func sessionStartKeepsLock(ctx *analysis.Context, start *syntax.FuncCall) bool {
	if start.Args == nil || len(start.Args.Args) == 0 {
		return true
	}
	options := semanticquery.NativeArray(ctx, semanticquery.CallArgument(start.Args, 0, "options"))
	if options == nil {
		return false
	}
	for _, item := range options.Items {
		key, known := semanticquery.NativeString(ctx, item.Key)
		if !known {
			return false
		}
		if closes, known := semanticquery.NativeTruth(ctx, item.Value); key == "read_and_close" && (closes || !known) {
			return false
		}
	}
	return true
}

// sessionFilesHandler proves that the session starts with the builtin files
// handler: the last handler selection written before session_start must be an
// unconditional ini_set of session.save_handler to 'files'. A later ini write
// with another or unknown value, an ini write to an unknown option, ini_restore,
// session_set_save_handler and a session_module_name setter all discard the proof.
func sessionFilesHandler(ctx *analysis.Context, start *syntax.FuncCall) bool {
	var last *syntax.FuncCall
	files := false
	for _, c := range ctx.Flow().Calls(syntax.EnclosingVariableScope(start)) {
		call, ok := c.Node.(*syntax.FuncCall)
		if !ok || call.Span().Start >= start.Span().Start || (last != nil && call.Span().Start < last.Span().Start) {
			continue
		}
		switch semanticquery.NativeBuiltinName(ctx, call) {
		case "ini_set", "ini_alter", "ini_restore":
			key, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 0, "option"))
			if known && key != "session.save_handler" {
				continue
			}
			value, set := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 1, "value"))
			last, files = call, known && set && value == "files"
		case "session_set_save_handler":
			last, files = call, false
		case "session_module_name":
			if call.Args != nil && len(call.Args.Args) > 0 {
				last, files = call, false
			}
		}
	}
	return files && semanticquery.NativeDominates(last, start)
}
