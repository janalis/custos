// Package sessioncookieoptionssetafterstart implements the native SessionCookieOptionsSetAfterStart inspection.
package sessioncookieoptionssetafterstart

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

const message = "Configure session cookie parameters before starting the session."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SessionCookieOptionsSetAfterStart" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)

	if name != "session_set_cookie_params" {
		return
	}
	start := semanticquery.GlobalStateCall(ctx, c, "session")
	if start == nil || !semanticquery.NativeBuiltin(ctx, start, "session_start") {
		return
	}
	if sessionStartKeepsActive(ctx, start) && !sessionInactiveProven(ctx, c, start) {
		ctx.ReportNode(c, message)
	}
}

// sessionInactiveProven proves that no session is active where c runs: the
// conditions guarding it narrow the start call's result to false, or a
// session_status() call made after the start to anything but
// PHP_SESSION_ACTIVE.
func sessionInactiveProven(ctx *analysis.Context, c, start *syntax.FuncCall) bool {
	return semanticquery.NativeEarlierCallProves(ctx, c, func(x syntax.Expr) (types.Type, func(types.Type) bool) {
		test, ok := x.(*syntax.FuncCall)
		switch {
		case !ok:
		case test == start:
			return types.Bool, semanticquery.OnlyFalse
		case test.Span().Start > start.Span().End && semanticquery.NativeBuiltin(ctx, test, "session_status"):
			// PHP_SESSION_DISABLED, PHP_SESSION_NONE or PHP_SESSION_ACTIVE.
			return types.Ints(types.IntRange{Lo: 0, Hi: 2}), func(t types.Type) bool { return t.IntsOnly(0, 1) }
		}
		return types.Unknown, nil
	})
}

// sessionStartKeepsActive proves the start options leave the session open.
// read_and_close closes it again at once; unknown options prove nothing.
func sessionStartKeepsActive(ctx *analysis.Context, start *syntax.FuncCall) bool {
	options := semanticquery.CallArgument(start.Args, 0, "options")
	if options == nil {
		return start.Args == nil || len(start.Args.Args) == 0
	}
	entries, known := semanticquery.NativeArrayEntries(ctx, options)
	if !known {
		return false
	}
	flag, set := entries["s:read_and_close"]
	closing, known := semanticquery.NativeTruth(ctx, flag)
	return !set || (known && !closing)
}
