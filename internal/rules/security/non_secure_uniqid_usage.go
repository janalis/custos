package security

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// nonSecureUniqidUsage reports uniqid() used without more_entropy, directly
// or as a string callback.
type nonSecureUniqidUsage struct{}

func init() { register(nonSecureUniqidUsage{}) }

const uniqidMsg = "Pass more_entropy = true to uniqid() to reduce collisions."

const uniqidClosure = "function ($value) { return uniqid($value, true); }"

var uniqidCallbackIndex = map[string]int{
	"call_user_func":       0,
	"call_user_func_array": 0,
	"array_map":            0,
	"array_filter":         1,
	"array_reduce":         1,
	"array_walk":           1,
	"array_walk_recursive": 1,
}

func (nonSecureUniqidUsage) ID() string { return "NonSecureUniqidUsage" }

func (nonSecureUniqidUsage) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func (nonSecureUniqidUsage) Semantic() {}

func (r nonSecureUniqidUsage) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	name := ctx.GlobalFunctionName(call)
	if name == "uniqid" {
		r.direct(ctx, call)
		return
	}
	idx, ok := uniqidCallbackIndex[name]
	if !ok || util.ArgCount(call) < 2 { // D3, E3
		return
	}
	arg, ok := call.Args.Args[idx].(*syntax.Arg)
	if !ok || arg.Value == nil {
		return
	}
	lit := uniqidLiteral(ctx, arg.Value) // D4
	if lit == nil {
		return
	}
	val, ok := util.StringLiteralValue(lit.Raw)
	if !ok || !strings.EqualFold(strings.TrimPrefix(val, `\`), "uniqid") { // D5
		return
	}
	span := arg.Value.Span()
	if idx != 0 {
		// F2: array_filter/array_reduce/array_walk* pass a different argument
		// list than a single prefix, so a one-parameter closure would change
		// what uniqid() receives; report without a fix.
		ctx.Report(span, uniqidMsg)
		return
	}
	// A namespaced or imported uniqid() would capture a bare call in the closure.
	closure := strings.Replace(uniqidClosure, "uniqid(", util.QualifiedBuiltin(ctx, "uniqid", span.Start)+"(", 1)
	ctx.Report(span, uniqidMsg, analysis.Fix{
		Title: "Use a uniqid() closure with more_entropy",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: span, NewText: closure}}
		},
	})
}

func (nonSecureUniqidUsage) direct(ctx *analysis.Context, call *syntax.FuncCall) {
	// D1, E2: Check only calls this for calls resolving to the global
	// uniqid() (GlobalFunctionName); a call always has an argument list.
	positional, named, spread := 0, false, false
	var texts []string
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
		texts = append(texts, ctx.Text(a))
	}
	if positional >= 2 { // D2
		return
	}
	list := call.Args.Span()
	var fixes []analysis.Fix
	if !spread {
		var b strings.Builder
		b.WriteByte('(')
		switch {
		case named:
			b.WriteString(strings.Join(texts, ", "))
			b.WriteString(", more_entropy: true")
		case len(texts) == 0:
			b.WriteString("'', true")
		default:
			b.WriteString(texts[0])
			b.WriteString(", true")
		}
		b.WriteByte(')')
		newText := b.String()
		fixes = append(fixes, analysis.Fix{
			Title: "Pass more_entropy = true",
			Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: list, NewText: newText}} },
		})
	}
	ctx.ReportNode(call, uniqidMsg, fixes...)
}

// uniqidLiteral returns the string literal a callback argument resolves to.
func uniqidLiteral(ctx *analysis.Context, e syntax.Expr) *syntax.Literal {
	if lit, ok := e.(*syntax.Literal); ok {
		if lit.LitKind == syntax.LitString {
			return lit
		}
		return nil
	}
	var found *syntax.Literal
	count := 0
	for _, v := range util.DiscoverValues(ctx.Types(), e) {
		if lit, ok := v.(*syntax.Literal); ok && lit.LitKind == syntax.LitString {
			found = lit
			count++
		}
	}
	if count != 1 {
		return nil
	}
	return found
}
