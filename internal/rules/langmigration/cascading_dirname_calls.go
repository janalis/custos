package langmigration

import (
	"strconv"
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// cascadingDirnameCalls collapses nested dirname() calls into one call with
// a levels argument.
type cascadingDirnameCalls struct{}

func init() { register(cascadingDirnameCalls{}) }

func (cascadingDirnameCalls) ID() string { return "CascadingDirnameCalls" }

func (cascadingDirnameCalls) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

// dirnameArgs returns the arguments of an eligible dirname call (any case,
// resolving to the global function; 1 or 2 positional arguments, no spread
// or named arguments).
func dirnameArgs(ctx *analysis.Context, e syntax.Node) ([]syntax.Expr, bool) {
	call, ok := util.IsFuncNamedFold(e, "dirname")
	if !ok || ctx.GlobalFunctionName(call) != "dirname" {
		return nil, false
	}
	args, ok := util.CallArgValues(call)
	if !ok || len(args) < 1 || len(args) > 2 {
		return nil, false
	}
	return args, true
}

func (cascadingDirnameCalls) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpver.PHP70 { // E1
		return
	}
	top := n.(*syntax.FuncCall)
	args, ok := dirnameArgs(ctx, top) // D1
	if !ok {
		return
	}
	if outer := directArgOf(top); outer != nil { // D2: only chain members
		if oargs, ok := dirnameArgs(ctx, outer); ok && oargs[0] == syntax.Expr(top) {
			return
		}
	}
	// D3: walk the chain.
	count := 0
	var levels []syntax.Expr
	var path syntax.Expr
	for cur := syntax.Expr(top); ; {
		a, ok := dirnameArgs(ctx, cur)
		if !ok {
			break
		}
		path = a[0]
		if len(a) == 1 {
			count++
		} else {
			levels = append(levels, a[1])
		}
		if _, isCall := a[0].(*syntax.FuncCall); !isCall {
			break
		}
		cur = a[0]
	}
	if path == args[0] { // D4
		return
	}
	// D5
	var exprs []string
	for _, l := range levels {
		if v, ok := parseLevel(ctx.Text(l)); ok {
			count += v
			continue
		}
		t := ctx.Text(l)
		if !isPrimaryLevel(l) { // spec Divergences: keep precedence
			t = "(" + t + ")"
		}
		exprs = append(exprs, t)
	}
	if count == 1 && len(exprs) == 0 { // D6
		return
	}
	// A namespaced dirname() would capture a bare call.
	fn := util.QualifiedBuiltin(ctx, "dirname", top.Span().Start) + "("
	repl := fn + ctx.Text(path) + ", " + strings.Join(append([]string{strconv.Itoa(count)}, exprs...), " + ") + ")" // D7
	span := top.Span()
	ctx.Report(span, "Collapse the nested dirname() calls into '"+repl+"'.", analysis.Fix{
		Title: "Replace with '" + repl + "'",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: span, NewText: repl}}
		},
	})
}

// parseLevel parses a signed 32-bit decimal integer. A literal with a
// leading zero followed by more digits (octal in PHP) is not parsed (spec
// Divergences).
func parseLevel(s string) (int, bool) {
	digits := strings.TrimLeft(s, "+-")
	if len(s)-len(digits) > 1 || digits == "" {
		return 0, false
	}
	if len(digits) > 1 && digits[0] == '0' {
		return 0, false
	}
	for i := 0; i < len(digits); i++ {
		if digits[i] < '0' || digits[i] > '9' {
			return 0, false
		}
	}
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, false
	}
	return int(v), true
}

// isPrimaryLevel reports whether a level expression can be joined with ` + `
// without parentheses.
func isPrimaryLevel(e syntax.Expr) bool {
	switch e.(type) {
	case *syntax.Variable, *syntax.ConstFetch, *syntax.ClassConstFetch, *syntax.Literal,
		*syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall, *syntax.PropertyFetch,
		*syntax.StaticPropertyFetch, *syntax.ArrayDimFetch, *syntax.Paren:
		return true
	}
	return false
}
