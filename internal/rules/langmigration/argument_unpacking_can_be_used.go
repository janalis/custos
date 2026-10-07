package langmigration

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// argumentUnpackingCanBeUsed reports call_user_func_array('fn', $args) that
// can be written as fn(...$args).
type argumentUnpackingCanBeUsed struct{}

func init() { register(argumentUnpackingCanBeUsed{}) }

func (argumentUnpackingCanBeUsed) ID() string { return "ArgumentUnpackingCanBeUsed" }

func (argumentUnpackingCanBeUsed) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func (argumentUnpackingCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpver.PHP56 { // D1
		return
	}
	call := n.(*syntax.FuncCall)
	if !ctx.IsGlobalFunctionCall(call, "call_user_func_array") {
		return // D2: any case; not a namespaced function of the same name
	}
	args, ok := util.CallArgValues(call)
	if !ok || len(args) != 2 { // D3
		return
	}
	lit, ok := args[0].(*syntax.Literal) // D4
	if !ok || lit.LitKind != syntax.LitString || strings.HasPrefix(lit.Raw, "<<<") {
		return
	}
	fn, _ := util.StringLiteralValue(lit.Raw) // D6
	if !isFunctionPath(fn) {
		return // not a plain function name (spec Divergences)
	}
	switch args[1].(type) { // D5
	case *syntax.Variable, *syntax.PropertyFetch, *syntax.StaticPropertyFetch, *syntax.Array,
		*syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
	default:
		return
	}
	keys := keysListSafe(ctx, args[1]) // D8
	if keys == keysString {
		return
	}
	repl := util.StringCallableFunction(ctx, fn, call.Span().Start) + "(..." + ctx.Text(args[1]) + ")" // D7
	span := call.Span()
	msg := "Call '" + repl + "' directly using argument unpacking (wrap with array_values() when keys are not sequential)."
	if keys == keysUnknown { // F2
		ctx.Report(span, msg)
		return
	}
	ctx.Report(span, msg, analysis.Fix{
		Title: "Replace with '" + repl + "'",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: span, NewText: repl}}
		},
	})
}

type unpackKeys int

const (
	keysSafe    unpackKeys = iota // unpacking cannot fail on string keys
	keysUnknown                   // the array may have string keys
	keysString                    // the array literal has a string key
)

// keysListSafe classifies the second argument (D8): from PHP 8.0 both forms
// treat string keys as named arguments, so any array is safe; below 8.0
// unpacking a string-keyed array throws, so only array literals whose keys
// are all absent or integer literals are known to be safe.
func keysListSafe(ctx *analysis.Context, arg syntax.Expr) unpackKeys {
	if ctx.PHP >= phpver.PHP80 {
		return keysSafe
	}
	arr, ok := arg.(*syntax.Array)
	if !ok {
		return keysUnknown
	}
	res := keysSafe
	for _, it := range arr.Items {
		if it == nil || it.Key == nil {
			continue
		}
		lit, ok := it.Key.(*syntax.Literal)
		switch {
		case ok && lit.LitKind == syntax.LitInt:
		case ok && lit.LitKind == syntax.LitString:
			return keysString
		default:
			res = keysUnknown
		}
	}
	return res
}

// isFunctionPath reports whether s is an optionally \-qualified,
// \-separated path of PHP identifiers: only such strings can be written as a
// direct call (`Cls::m`, the empty string or names with spaces cannot).
func isFunctionPath(s string) bool {
	s = strings.TrimPrefix(s, `\`)
	for _, seg := range strings.Split(s, `\`) {
		if seg == "" || (seg[0] >= '0' && seg[0] <= '9') {
			return false
		}
		for i := 0; i < len(seg); i++ {
			c := seg[i]
			if !(c == '_' || c >= 0x80 || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
				return false
			}
		}
	}
	return true
}
