// Package nocontentresponsewithbody implements the native NoContentResponseWithBody inspection.
package nocontentresponsewithbody

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Omit response content for status 204."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "NoContentResponseWithBody" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KEcho} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	check(ctx, n, message)
}

func check(ctx *analysis.Context, n syntax.Node, message string) {
	echo := n.(*syntax.Echo)
	call := semanticquery.GlobalStateCall(ctx, echo, "response")
	if call == nil {
		return
	}
	status, known := semanticquery.NativeInt(ctx, semanticquery.CallArgument(call.Args, 0, "response_code"))
	if !known || status != 204 || noContentStatusReplaced(ctx, call, echo) {
		return
	}
	// Output written into an open buffer may still be discarded.
	if _, buffering := ctx.Flow().GlobalStateBefore(echo, "output-buffer"); buffering {
		return
	}
	for _, expr := range echo.Exprs {
		if content, known := semanticquery.NativeString(ctx, expr); known && content != "" {
			ctx.ReportNode(echo, message)
			return
		}
	}
}

// noContentStatusReplaced reports whether a header() call that may replace the
// status set by the given http_response_code call can run before the echo: one
// placed between them, or later in a loop that repeats the echo.
func noContentStatusReplaced(ctx *analysis.Context, set *syntax.FuncCall, echo *syntax.Echo) bool {
	scope := syntax.EnclosingVariableScope(echo)
	limit := echo.Span().Start
	for p := echo.Parent(); p != nil && p != scope; p = p.Parent() {
		switch p.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			limit = p.Span().End
		}
	}
	for _, c := range ctx.Flow().Calls(scope) {
		header, ok := c.Node.(*syntax.FuncCall)
		if !ok || header.Span().Start < set.Span().End || header.Span().Start >= limit {
			continue
		}
		if semanticquery.NativeBuiltinName(ctx, header) == "header" && !noContentHeaderKeepsStatus(ctx, header) {
			return true
		}
	}
	return false
}

// noContentHeaderKeepsStatus proves that a header() call leaves a 204 status
// in place: no other explicit status argument, and a field line that is
// neither a status line nor a Location or Status field.
func noContentHeaderKeepsStatus(ctx *analysis.Context, c *syntax.FuncCall) bool {
	explicit, state := semanticquery.CallArgumentBinding(c.Args, 2, "response_code")
	if state == semanticquery.ArgumentUnknown {
		return false
	}
	if state == semanticquery.ArgumentBound {
		code, known := semanticquery.NativeInt(ctx, explicit)
		if !known {
			return false
		}
		if code != 0 {
			return code == 204
		}
	}
	line := syntax.UnwrapParens(semanticquery.CallArgument(c.Args, 0, "header"))
	text, complete := semanticquery.NativeString(ctx, line)
	if !complete {
		// A concatenation still exposes the field name in its leading literal.
		for {
			concat, ok := line.(*syntax.Binary)
			if !ok || concat.Op.Kind != syntax.TDot {
				break
			}
			line = syntax.UnwrapParens(concat.Left)
		}
		text, _ = semanticquery.NativeString(ctx, line)
	}
	text = strings.ToLower(text)
	if strings.HasPrefix(text, "http/") {
		fields := strings.Fields(text)
		return complete && len(fields) > 1 && fields[1] == "204"
	}
	name, _, named := strings.Cut(text, ":")
	if !named {
		return complete
	}
	name = strings.TrimSpace(name)
	return name != "location" && name != "status"
}
