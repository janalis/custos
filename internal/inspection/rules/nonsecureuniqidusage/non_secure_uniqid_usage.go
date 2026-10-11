package nonsecureuniqidusage

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// nonSecureUniqidUsage reports uniqid() used without more_entropy, directly
// or as a string callback.
type nonSecureUniqidUsage struct{}

const (
	uniqidMsg     = "Pass more_entropy = true to uniqid() to reduce collisions."
	uniqidClosure = "function ($value = '') { return uniqid((string) $value, true); }"
)

var uniqidCallbackIndex = map[string]int{
	"call_user_func":       0,
	"call_user_func_array": 0,
	"array_map":            0,
	"array_filter":         1,
	"array_reduce":         1,
	"array_walk":           1,
	"array_walk_recursive": 1,
}

func (nonSecureUniqidUsage) ID() string               { return "NonSecureUniqidUsage" }
func (nonSecureUniqidUsage) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (nonSecureUniqidUsage) Semantic()                {}
func (r nonSecureUniqidUsage) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	name := ctx.GlobalFunctionName(call)
	if name == "uniqid" {
		r.direct(ctx, call)
		return
	}
	idx, ok := uniqidCallbackIndex[name]
	if !ok || astquery.ArgCount(call) < 2 { // D3, E3
		return
	}
	// plain: every argument is positional and none is unpacked, so the
	// argument list alone tells what uniqid() receives.
	plain, positional := true, 0
	var cb *syntax.Arg
	var list syntax.Expr // call_user_func_array's argument array
	for i, a := range call.Args.Args {
		arg, ok := a.(*syntax.Arg)
		if !ok {
			return // first-class callable syntax
		}
		switch {
		case arg.Unpack:
			plain = false
		case arg.Name != nil:
			plain = false
			switch arg.Name.Value {
			case "callback":
				cb = arg
			case "args":
				list = arg.Value
			case "more_entropy":
				if name == "call_user_func" {
					return // E4: forwarded by name
				}
			}
		default:
			positional++
			switch i {
			case idx:
				cb = arg
			case 1:
				list = arg.Value
			}
		}
	}
	if cb == nil || cb.Value == nil {
		return
	}
	lit, literal := uniqidLiteral(ctx, cb.Value) // D4
	if lit == nil {
		return
	}
	val, ok := astquery.StringLiteralValue(lit.Raw)
	if !ok || !strings.EqualFold(strings.TrimPrefix(val, `\`), "uniqid") { // D5
		return
	}
	// F2 needs the callback written as a literal (a variable may hold another
	// callable on a path value discovery does not see) and exactly one
	// positional value forwarded as the prefix.
	fixable := idx == 0 && literal && plain && positional == 2
	switch name {
	case "call_user_func", "array_map":
		if positional >= 3 {
			return // E4: the second forwarded value is more_entropy
		}
	case "call_user_func_array":
		entropy, single := forwardedList(list)
		if entropy {
			return // E4
		}
		fixable = fixable && single
	}
	span := cb.Value.Span()
	if !fixable {
		// F2: array_filter/array_reduce/array_walk* pass a different argument
		// list than a single prefix, and the other shapes above cannot be
		// proven to forward exactly one positional value; report without a fix.
		ctx.Report(span, uniqidMsg)
		return
	}
	// A namespaced or imported uniqid() would capture a bare call in the closure.
	closure := strings.Replace(uniqidClosure, "uniqid(", semanticquery.QualifiedBuiltin(ctx, "uniqid", span.Start)+"(", 1)
	ctx.Report(span, uniqidMsg, diagnostic.Fix{
		Title: "Use a uniqid() closure with more_entropy",
		Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: span, NewText: closure}}
		},
	})
}

// forwardedList inspects the argument array of call_user_func_array: entropy
// when a literal array already supplies more_entropy (a second element or a
// 'more_entropy' key), single when it is a literal list of exactly one plain
// element. Both are false for anything that is not an array literal.
func forwardedList(e syntax.Expr) (entropy, single bool) {
	arr, ok := e.(*syntax.Array)
	if !ok {
		return false, false
	}
	n := 0
	single = len(arr.Items) == 1
	for _, it := range arr.Items {
		if it == nil || it.Unpack {
			single = false
			continue
		}
		n++
		if it.Key != nil || it.ByRef {
			single = false
		}
		if key, ok := it.Key.(*syntax.Literal); ok && key.LitKind == syntax.LitString {
			if v, _ := astquery.StringLiteralValue(key.Raw); v == "more_entropy" {
				return true, false
			}
		}
	}
	return n >= 2, single
}

func (nonSecureUniqidUsage) direct(ctx *analysis.Context, call *syntax.FuncCall) {
	// D1, E2: Check only calls this for calls resolving to the global
	// uniqid() (GlobalFunctionName); a call always has an argument list.
	positional, named, spread := 0, false, false
	for _, a := range call.Args.Args {
		arg, ok := a.(*syntax.Arg)
		if !ok {
			return // first-class callable syntax
		}
		switch {
		case arg.Unpack:
			spread = true
		case arg.Name != nil:
			named = true
			if arg.Name.Value == "more_entropy" {
				return // E1
			}
		default:
			if !named && !spread {
				positional++
			}
		}
	}
	if positional >= 2 { // D2
		return
	}
	var fixes []diagnostic.Fix
	if !spread {
		// F1: insert rather than rebuild the list, so comments, spacing and a
		// trailing comma inside the parentheses survive.
		at, text := call.Args.Span().Start+1, "'', true"
		if args := call.Args.Args; len(args) > 0 {
			at, text = args[len(args)-1].Span().End, ", true"
			if named {
				text = ", more_entropy: true"
			}
		}
		fixes = append(fixes, diagnostic.Fix{
			Title: "Pass more_entropy = true",
			Edits: func() []diagnostic.TextEdit {
				return []diagnostic.TextEdit{{Span: syntax.Span{Start: at, End: at}, NewText: text}}
			},
		})
	}
	ctx.ReportNode(call, uniqidMsg, fixes...)
}

// uniqidLiteral returns the string literal a callback argument resolves to,
// and whether the argument is that literal itself.
func uniqidLiteral(ctx *analysis.Context, e syntax.Expr) (*syntax.Literal, bool) {
	if lit, ok := e.(*syntax.Literal); ok {
		if lit.LitKind == syntax.LitString {
			return lit, true
		}
		return nil, false
	}
	var found *syntax.Literal
	count := 0
	for _, v := range semanticquery.DiscoverValues(ctx.Types(), e) {
		if lit, ok := v.(*syntax.Literal); ok && lit.LitKind == syntax.LitString {
			found = lit
			count++
		}
	}
	if count != 1 {
		return nil, false
	}
	return found, false
}
