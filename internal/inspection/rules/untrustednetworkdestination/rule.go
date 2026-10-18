// Package untrustednetworkdestination implements the native UntrustedNetworkDestination inspection.
package untrustednetworkdestination

import (
	"regexp"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/flow"
)

const message = "Validate the outbound network destination and redirects."

var (
	// fixedOrigin matches literal text that already spells a complete
	// scheme and authority: the delimiter after the host is what stops a
	// later operand from extending or replacing it (`https://host` + `@evil`).
	fixedOrigin = regexp.MustCompile(`^(?i)[a-z][a-z0-9+.\-]*://[^/?#\\\s]+[/?#]`)
	// networkScheme matches literal text selecting a remote stream wrapper.
	networkScheme = regexp.MustCompile(`^(?i)(?:https?|ftps?)://`)
)

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule { return rule{} }
func (rule) ID() string  { return "UntrustedNetworkDestination" }
func (rule) Semantic()   {}
func (rule) Flow()       {}
func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall, syntax.KMethodCall, syntax.KStaticCall}
}

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c, function := n.(*syntax.FuncCall)
	if !function {
		if wrapperSink(ctx, n) {
			ctx.ReportNode(n, message)
		}
		return
	}
	report := false
	switch semanticquery.NativeBuiltinName(ctx, c) {
	case "curl_exec":
		report = curlDestination(ctx, c)
	case "file_get_contents", "fopen", "readfile", "file":
		report = streamDestination(ctx, semanticquery.CallArgument(c.Args, 0, "filename"))
	case "copy":
		report = streamDestination(ctx, semanticquery.CallArgument(c.Args, 0, "from"))
	}
	if report || wrapperSink(ctx, c) {
		ctx.ReportNode(c, message)
	}
}

// curlDestination replays the URL configured on the executed handle: the
// curl_init argument, then ordered possible CURLOPT_URL assignments. A
// mandatory assignment replaces the state; a conditional one joins it.
func curlDestination(ctx *analysis.Context, exec *syntax.FuncCall) bool {
	handle := semanticquery.CallArgument(exec.Args, 0, "handle")
	report := false
	if origin, ok := semanticquery.NativeValue(ctx, handle).(*syntax.FuncCall); ok && semanticquery.NativeBuiltin(ctx, origin, "curl_init") {
		report = untrusted(ctx, semanticquery.CallArgument(origin.Args, 0, "url"))
	}
	history := semanticquery.NativeCallHistory(ctx, exec, handle, "curl_setopt", "curl_setopt_array")
	if !history.Complete {
		return false
	}
	for _, event := range history.Events {
		option := event.Node.(*syntax.FuncCall)
		switch semanticquery.NativeBuiltinName(ctx, option) {
		case "curl_setopt":
			if curlURLOption(ctx, semanticquery.CallArgument(option.Args, 1, "option")) {
				unsafe := untrusted(ctx, semanticquery.CallArgument(option.Args, 2, "value"))
				if event.Must {
					report = unsafe
				} else {
					report = report || unsafe
				}
			}
		case "curl_setopt_array":
			// Only a literal option array has known keys; anything else
			// leaves the destination proven so far untouched.
			if options := semanticquery.NativeArray(ctx, semanticquery.CallArgument(option.Args, 1, "options")); options != nil {
				for _, item := range options.Items {
					if curlURLOption(ctx, item.Key) {
						unsafe := untrusted(ctx, item.Value)
						if event.Must {
							report = unsafe
						} else {
							report = report || unsafe
						}
					}
				}
			}
		}
	}
	return report
}

// curlURLOption resolves the option constant, so `\CURLOPT_URL` and a
// namespace fallback match while a namespaced constant of that name does not.
func curlURLOption(ctx *analysis.Context, e syntax.Expr) bool {
	constant, ok := syntax.UnwrapParens(e).(*syntax.ConstFetch)
	return ok && semanticquery.GlobalConstName(ctx, constant) == "CURLOPT_URL"
}

// streamDestination reports a filename read through a remote stream wrapper
// the request chooses: either the request supplies the start of the name
// (and with it the scheme), or a literal network scheme is followed by a
// request-derived host. Names starting with other fixed text stay local.
func streamDestination(ctx *analysis.Context, target syntax.Expr) bool {
	if !untrusted(ctx, target) {
		return false
	}
	lead := leading(ctx, target)
	if networkScheme.MatchString(literalText(lead)) {
		return true
	}
	return ctx.Flow().Tainted(lead, flow.URL) && !schemeless(ctx, lead)
}

// schemeless recognizes leading operands whose result cannot contain the
// `://` a stream wrapper needs.
func schemeless(ctx *analysis.Context, e syntax.Expr) bool {
	switch x := e.(type) {
	case *syntax.Unary:
		return x.Op.Kind == syntax.TIntCast || x.Op.Kind == syntax.TDoubleCast || x.Op.Kind == syntax.TBoolCast
	case *syntax.FuncCall:
		switch semanticquery.NativeBuiltinName(ctx, x) {
		case "intval", "floatval", "basename", "urlencode", "rawurlencode":
			return true
		}
	}
	return false
}

// untrusted reports a request-derived destination whose origin is not fixed.
func untrusted(ctx *analysis.Context, url syntax.Expr) bool {
	return ctx.Flow().Tainted(url, flow.URL) && !fixedDestination(ctx, url)
}

// fixedDestination proves the origin is spelled by the source: the leading
// literal contains scheme, host and the delimiter ending the authority, or
// the whole value is a match whose every arm is a literal or such a
// concatenation.
func fixedDestination(ctx *analysis.Context, e syntax.Expr) bool {
	e = syntax.UnwrapParens(e)
	if v, ok := e.(*syntax.Variable); ok && !aliased(ctx, v) {
		if value := semanticquery.NativeValue(ctx, e); value != nil {
			e = value
		}
	}
	if selection, ok := e.(*syntax.Match); ok {
		return fixedArms(selection)
	}
	return fixedOrigin.MatchString(literalText(leading(ctx, e)))
}

// fixedArms stays syntactic: arm results are not resolved through variables.
func fixedArms(selection *syntax.Match) bool {
	for _, arm := range selection.Arms {
		result := syntax.UnwrapParens(arm.Body)
		if nested, ok := result.(*syntax.Match); ok {
			if !fixedArms(nested) {
				return false
			}
			continue
		}
		if _, throws := result.(*syntax.Throw); throws {
			continue // no destination at all
		}
		if _, literal := astquery.QuotedStringValue(result); literal {
			continue
		}
		for {
			concat, ok := result.(*syntax.Binary)
			if !ok || concat.Op.Kind != syntax.TDot {
				break
			}
			result = syntax.UnwrapParens(concat.Left)
		}
		if !fixedOrigin.MatchString(literalText(result)) {
			return false
		}
	}
	return true
}

// leading returns the operand supplying the first bytes of a string
// expression, looking through parentheses, concatenations, interpolation and
// variables with a single reaching value.
func leading(ctx *analysis.Context, e syntax.Expr) syntax.Expr {
	for steps := 0; ; steps++ {
		e = syntax.UnwrapParens(e)
		switch x := e.(type) {
		case *syntax.Binary:
			if x.Op.Kind == syntax.TDot {
				e = x.Left
				continue
			}
		case *syntax.Assign:
			// `$u .= x` leaves `$u . x`: its first bytes are those `$u` held.
			if x.Op.Kind == syntax.TConcatEqual {
				e = x.Var
				continue
			}
		case *syntax.InterpolatedString:
			if !x.Backtick && len(x.Parts) > 0 {
				e = x.Parts[0]
				continue
			}
		case *syntax.Variable:
			if value := semanticquery.NativeValue(ctx, x); value != nil && value != e && steps < 64 && !aliased(ctx, x) {
				e = value
				continue
			}
		}
		return e
	}
}

// literalText returns the text of a string literal or of a literal
// interpolation part, and "" for anything else.
func literalText(e syntax.Expr) string {
	if text, ok := astquery.QuotedStringValue(e); ok {
		return text
	}
	if part, ok := e.(*syntax.StringPart); ok {
		return part.Raw
	}
	return ""
}

// wrapperSink proves a request value reaches a curl destination inside a
// project function through its summarized sink contract.
func wrapperSink(ctx *analysis.Context, c syntax.Node) bool {
	for _, sink := range ctx.Flow().Sinks(c) {
		if sink.Name != "curl_init" && sink.Name != "curl_setopt" {
			continue
		}
		if sink.Context != flow.URL || !sink.Argument.Complete || sink.Argument.Safe&flow.URL != 0 {
			continue
		}
		for _, source := range sink.Argument.Sources {
			if source.Parameter < 0 {
				return true
			}
		}
	}
	return false
}
