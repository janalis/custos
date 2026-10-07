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
	if f := ctx.Types().ResolveFunction(call); f != nil && !strings.EqualFold(strings.TrimPrefix(f.FQN, `\`), "call_user_func_array") {
		return // D2: unqualified call resolving to a user function
	}
	args, ok := util.CallArgValues(call)
	if !ok || len(args) != 2 { // D3
		return
	}
	lit, ok := args[0].(*syntax.Literal) // D4
	if !ok || lit.LitKind != syntax.LitString || strings.HasPrefix(lit.Raw, "<<<") {
		return
	}
	fn, ok := util.StringLiteralValue(lit.Raw) // D6
	if !ok {
		return
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
