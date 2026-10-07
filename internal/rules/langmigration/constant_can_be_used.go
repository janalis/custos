package langmigration

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// constantCanBeUsed reports calls whose result PHP already exposes as a
// constant, version_compare(PHP_VERSION, ...) checks and PHP_OS sniffing.
type constantCanBeUsed struct{}

func init() { register(constantCanBeUsed{}) }

func (constantCanBeUsed) ID() string { return "ConstantCanBeUsed" }

func (constantCanBeUsed) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall, syntax.KConstFetch}
}

var constantForCall = map[string]string{
	"phpversion":    "PHP_VERSION",
	"php_sapi_name": "PHP_SAPI",
	"get_class":     "__CLASS__",
	"pi":            "M_PI",
}

var versionCompareOps = map[string]string{
	"<": "<", "lt": "<",
	"<=": "<=", "le": "<=",
	">": ">", "gt": ">",
	">=": ">=", "ge": ">=",
	"==": "===", "=": "===", "eq": "===",
	"!=": "!==", "<>": "!==", "ne": "!==",
}

func (r constantCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	switch n := n.(type) {
	case *syntax.FuncCall:
		r.checkCall(ctx, n)
	case *syntax.ConstFetch:
		r.checkOS(ctx, n)
	}
}

func (constantCanBeUsed) checkCall(ctx *analysis.Context, call *syntax.FuncCall) {
	name := ctx.GlobalFunctionName(call) // any case, global functions only
	if name == "" || call.Args == nil {
		return
	}
	span := call.Span()
	if c, ok := constantForCall[name]; ok { // D1
		if len(call.Args.Args) != 0 {
			return
		}
		if name == "get_class" && classScope(call) == nil { // E4: __CLASS__ is '' there
			return
		}
		if c != "__CLASS__" {
			c = util.QualifiedGlobalConst(ctx, c, span.Start) // a namespaced constant would capture a bare name
		}
		ctx.Report(span, "Use the "+c+" constant instead of this call.", analysis.Fix{
			Title: "Replace with " + c,
			Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: span, NewText: c}} },
		})
		return
	}
	if name != "version_compare" {
		return
	}
	args, ok := util.CallArgValues(call)
	if !ok || len(args) != 3 { // D2
		return
	}
	if cf, ok := args[0].(*syntax.ConstFetch); !ok || util.GlobalConstName(ctx, cf) != "PHP_VERSION" {
		return
	}
	ver, ok := util.QuotedStringContent(args[1])
	if !ok {
		return
	}
	v, ok := versionID(ver) // D3
	if !ok {
		return
	}
	opRaw, ok := util.QuotedStringContent(args[2])
	if !ok {
		return
	}
	op, ok := versionCompareOps[opRaw]
	if !ok {
		return
	}
	// E5: equality ignores the missing patch level only in PHP_VERSION_ID.
	if (op == "===" || op == "!==") && strings.Count(ver, ".") != 2 {
		return
	}
	// D4b: a short version sorts below its `.0` release, so PHP_VERSION is
	// never equal to it: `>` means `>= V`, `<=` means `< V`.
	if strings.Count(ver, ".") != 2 {
		switch op {
		case ">":
			op = ">="
		case "<=":
			op = "<"
		}
	}
	expr := util.QualifiedGlobalConst(ctx, "PHP_VERSION_ID", span.Start) + " " + op + " " + v // D4
	repl := expr
	if bindsTighterThanComparison(call) { // spec Divergences
		repl = "(" + expr + ")"
	}
	ctx.Report(span, "Replace with '"+expr+"'.", analysis.Fix{
		Title: "Replace with '" + expr + "'",
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: span, NewText: repl}} },
	})
}

// bindsTighterThanComparison reports whether the call is the operand of an
// operator that would capture part of an unparenthesised comparison.
func bindsTighterThanComparison(call *syntax.FuncCall) bool {
	switch p := call.Parent().(type) {
	case *syntax.Unary:
		return true
	case *syntax.Binary:
		switch p.Op.Kind {
		case syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr, syntax.TXor, syntax.TCoalesce:
			return false
		}
		return true
	}
	return false
}

// versionID converts `M[.m[.p]]` (single-digit M and m, any-digit p) into a
// PHP_VERSION_ID value.
func versionID(s string) (string, bool) {
	parts := strings.Split(s, ".")
	if len(parts) > 3 {
		return "", false
	}
	for i, p := range parts {
		if p == "" || (i < 2 && len(p) != 1) {
			return "", false
		}
		for j := 0; j < len(p); j++ {
			if p[j] < '0' || p[j] > '9' {
				return "", false
			}
		}
	}
	out := parts[0]
	for i := 1; i < 3; i++ {
		p := "0"
		if i < len(parts) {
			p = parts[i]
		}
		if len(p) == 1 {
			p = "0" + p
		}
		out += p
	}
	return out, true
}

var osSniffers = map[string]bool{
	"strpos": true, "stripos": true, "mb_strpos": true, "mb_stripos": true,
	"strncmp": true, "strncasecmp": true, "substr": true, "mb_substr": true,
}

var caseFolders = map[string]bool{
	"strtolower": true, "mb_strtolower": true, "strtoupper": true, "mb_strtoupper": true,
}

func (constantCanBeUsed) checkOS(ctx *analysis.Context, c *syntax.ConstFetch) {
	if ctx.PHP < phpver.PHP72 || util.GlobalConstName(ctx, c) != "PHP_OS" {
		return
	}
	f := util.ParentFuncCall(c) // D5
	if f == nil {
		return
	}
	fname := ctx.GlobalFunctionName(f)
	if !osSniffers[fname] {
		return
	}
	isSubstr := fname == "substr" || fname == "mb_substr"
	x := f // D6
	if isSubstr {
		if outer := util.ParentFuncCall(f); outer != nil {
			if caseFolders[ctx.GlobalFunctionName(outer)] {
				x = outer
			}
		}
	}
	bin, ok := x.Parent().(*syntax.Binary) // D7
	if !ok {
		return
	}
	switch bin.Op.Kind {
	case syntax.TIsEqual, syntax.TIsIdentical, syntax.TIsNotIdentical:
	case syntax.TIsNotEqual:
		if ctx.SpanText(bin.Op.Span) != "!=" {
			return
		}
	default:
		return
	}
	other := bin.Left
	if bin.Left == syntax.Expr(x) {
		other = bin.Right
	}
	if isSubstr { // D8
		if lit, ok := other.(*syntax.Literal); !ok || lit.LitKind != syntax.LitString {
			return
		}
	} else if !isNumberOrFalse(other) {
		return
	}
	ctx.ReportNode(x, "Compare PHP_OS_FAMILY instead of sniffing PHP_OS.")
}

func isNumberOrFalse(e syntax.Expr) bool {
	if u, ok := e.(*syntax.Unary); ok && u.Op.Kind == syntax.TMinus {
		e = u.Expr
	}
	if lit, ok := e.(*syntax.Literal); ok {
		return lit.LitKind == syntax.LitInt || lit.LitKind == syntax.LitFloat
	}
	if c, ok := e.(*syntax.ConstFetch); ok && c.Name != nil {
		return strings.EqualFold(c.Name.Value, "false")
	}
	return false
}
