package variablefunctionsusage

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// variableFunctionsUsage reports call_user_func()-style calls whose callable
// or argument list is known at the call site and can be written directly.
type variableFunctionsUsage struct{}

// Semantic marks the rule as needing the project index (symbols or types
// declared in other files).
func (variableFunctionsUsage) Semantic()                {}
func (variableFunctionsUsage) ID() string               { return "VariableFunctionsUsage" }
func (variableFunctionsUsage) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func (variableFunctionsUsage) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	name := ctx.GlobalFunctionName(call) // D1: the global function, any case
	if name == "" || call.Args == nil {
		return
	}
	var repl, msg string
	var ok bool
	switch name {
	case "call_user_func_array", "forward_static_call_array":
		repl, ok = vfuInlineArgs(ctx, call, semanticquery.QualifiedBuiltin(ctx, strings.TrimSuffix(name, "_array"), call.Span().Start))
		msg = "Pass the arguments inline: '" + repl + "'."
	case "call_user_func", "forward_static_call":
		repl, ok = vfuDirectCall(ctx, call, name == "forward_static_call")
		msg = "Call it directly: '" + repl + "'."
	default:
		return
	}
	if !ok {
		return
	}
	span := call.Span()
	ctx.Report(span, msg, diagnostic.Fix{
		Title: "Replace with '" + repl + "'",
		Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: span, NewText: repl}} },
	})
}

// vfuArrayValues returns the element values of an array literal; ok is false
// when an element is by-reference, a spread or an empty slot.
func vfuArrayValues(arr *syntax.Array) ([]syntax.Expr, bool) {
	out := make([]syntax.Expr, 0, len(arr.Items))
	for _, it := range arr.Items {
		if it == nil || it.Value == nil || it.ByRef || it.Unpack {
			return nil, false
		}
		out = append(out, it.Value)
	}
	return out, true
}

// vfuInlineArgs implements Part A (D1–D4, F1).
func vfuInlineArgs(ctx *analysis.Context, call *syntax.FuncCall, target string) (string, bool) {
	list := call.Args.Args
	if len(list) != 2 {
		return "", false
	}
	a0, ok0 := list[0].(*syntax.Arg)
	a1, ok1 := list[1].(*syntax.Arg)
	if !ok0 || !ok1 || a1.Unpack || a1.Name != nil {
		return "", false
	}
	arr, ok := a1.Value.(*syntax.Array)
	if !ok {
		return "", false
	}
	vals, ok := vfuArrayValues(arr)
	if !ok || len(vals) == 0 {
		return "", false
	}
	if ctx.PHP >= phpversion.PHP80 { // D4a: non-integer keys are named arguments
		for _, it := range arr.Items {
			if it.Key == nil {
				continue
			}
			if lit, ok := it.Key.(*syntax.Literal); !ok || lit.LitKind != syntax.LitInt {
				return "", false
			}
		}
	}
	parts := []string{ctx.Text(a0)}
	for _, v := range vals {
		parts = append(parts, ctx.Text(v))
	}
	return target + "(" + strings.Join(parts, ", ") + ")", true
}

// vfuDirectCall implements Part B (D5–D8).
func vfuDirectCall(ctx *analysis.Context, call *syntax.FuncCall, static bool) (string, bool) {
	list := call.Args.Args
	if len(list) == 0 {
		return "", false
	}
	a0arg, ok := list[0].(*syntax.Arg)
	if !ok || a0arg.Value == nil || a0arg.Unpack {
		return "", false
	}
	a0 := a0arg.Value
	for _, a := range list[1:] { // D5a: call-time `&` is a fatal error
		if arg, ok := a.(*syntax.Arg); !ok || arg.ByRef { // or a misplaced `...`
			return "", false
		}
	}
	var first, second syntax.Expr
	if arr, ok := a0.(*syntax.Array); ok { // D6a
		var vals []syntax.Expr
		for _, it := range arr.Items {
			if it != nil && it.Value != nil {
				vals = append(vals, it.Value)
			}
		}
		if len(vals) == 0 {
			return "", false
		}
		first = vals[0]
		if len(vals) > 1 {
			second = vals[1]
		}
		if _, isVar := astquery.PlainVariableName(first); isVar {
			t := ctx.TypeOf(first)
			if t.IsUnknown() || t.HasAny("mixed", "string") {
				return "", false
			}
		}
	} else { // D6b
		if _, isVar := astquery.PlainVariableName(a0); isVar && ctx.PHP < phpversion.PHP54 {
			return "", false
		}
		first = a0
	}

	args := vfuArgsText(ctx, list[1:]) // D8
	_, firstIsVar := astquery.PlainVariableName(first)
	_, _, firstIsStr := astquery.QuotedStringRaw(first)
	if !firstIsVar && !firstIsStr { // E5: covers the `:` override too
		return "", false
	}
	var method string
	if second != nil { // D7
		method = "{" + ctx.Text(second) + "}"
		if _, _, ok := astquery.QuotedStringRaw(second); ok {
			c, _ := astquery.StringLiteralValue(second.(*syntax.Literal).Raw)
			if vfuParentCallable(c) {
				return "", false
			}
			if strings.Contains(c, ":") {
				if cls, m, ok := strings.Cut(c, "::"); ok { // F2: absolute class name
					c = semanticquery.StringCallableClass(ctx, cls, call.Span().Start) + "::" + m
				}
				return c + "(" + args + ")", true
			}
			method = c
		} else if _, isVar := astquery.PlainVariableName(second); isVar {
			method = ctx.Text(second)
		}
	}
	target := ctx.Text(first)
	at := call.Span().Start
	if firstIsStr {
		target, _ = astquery.StringLiteralValue(first.(*syntax.Literal).Raw)
		if second == nil && vfuParentCallable(target) { // E4 for a single string
			return "", false
		}
		// F2: string callables hold absolute names; spell them so the
		// direct call reaches the same function or class.
		if second != nil {
			target = semanticquery.StringCallableClass(ctx, target, at)
		} else if cls, m, ok := strings.Cut(target, "::"); ok {
			target = semanticquery.StringCallableClass(ctx, cls, at) + "::" + m
		} else {
			// D7a: a global name shadowed at the call (namespace function or
			// `use function`) is reached on purpose through the string.
			if fn := strings.TrimPrefix(target, `\`); fn != "" && !strings.Contains(fn, `\`) && !semanticquery.BareReachesGlobal(ctx, fn, at) {
				return "", false
			}
			target = semanticquery.StringCallableFunction(ctx, target, at)
		}
	}
	if second == nil {
		return target + "(" + args + ")", true
	}
	sep := "->"
	if !firstIsVar || static {
		sep = "::"
	}
	return target + sep + method + "(" + args + ")", true
}

// vfuArgsText renders arguments verbatim, keeping spreads and named-argument
// labels (call-time `&` arguments are rejected earlier, D5a).
func vfuArgsText(ctx *analysis.Context, list []syntax.Expr) string {
	parts := make([]string, 0, len(list))
	for _, a := range list {
		arg := a.(*syntax.Arg) // placeholders are rejected earlier (D5a)
		var b strings.Builder
		if arg.Name != nil {
			b.WriteString(arg.Name.Value + ": ")
		}
		if arg.Unpack {
			b.WriteString("...")
		}
		b.WriteString(ctx.Text(arg.Value))
		parts = append(parts, b.String())
	}
	return strings.Join(parts, ", ")
}

// vfuParentCallable reports a callable string naming a parent:: method
// (`parent` is case-insensitive in PHP).
func vfuParentCallable(s string) bool {
	return len(s) >= 8 && strings.EqualFold(s[:8], "parent::")
}
