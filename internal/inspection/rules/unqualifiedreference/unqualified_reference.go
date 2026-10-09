package unqualifiedreference

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// unqualifiedReference reports unqualified references to global functions
// and constants inside a namespace (and unqualified string callbacks) that
// a leading backslash would let PHP bind at compile time.
type unqualifiedReference struct{}

// Semantic marks the rule as needing the project index (symbols or types
// declared in other files).
func (unqualifiedReference) Semantic()  {}
func (unqualifiedReference) ID() string { return "UnqualifiedReference" }
func (unqualifiedReference) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall, syntax.KConstFetch}
}

var uqOpcodeFunctions = setOf(
	"array_slice", "assert", "boolval", "call_user_func", "call_user_func_array",
	"chr", "count", "defined", "doubleval", "floatval", "func_get_args",
	"func_num_args", "get_called_class", "get_class", "gettype", "in_array",
	"intval", "is_array", "is_bool", "is_double", "is_float", "is_int",
	"is_integer", "is_long", "is_null", "is_object", "is_real", "is_resource",
	"is_string", "ord", "strlen", "strval", "function_exists", "is_callable",
	"extension_loaded", "dirname", "constant", "define", "array_key_exists",
	"is_scalar", "sizeof", "ini_get", "sprintf", "printf",
)

var uqSkippedConstants = setOf(
	"true", "false", "null", "__line__", "__file__", "__dir__", "__function__",
	"__class__", "__trait__", "__method__", "__namespace__",
)

// uqSkippedConstant implements D2: true/false/null and the magic constants,
// which PHP matches case-insensitively.
func uqSkippedConstant(name string) bool { return uqSkippedConstants[strings.ToLower(name)] }

var uqCallbackPositions = map[string]int{
	"call_user_func": 0, "call_user_func_array": 0, "array_map": 0,
	"array_filter": 1, "array_walk": 1, "array_reduce": 1,
}

func setOf(items ...string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, s := range items {
		m[s] = true
	}
	return m
}

func (r unqualifiedReference) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP70 { // E1
		return
	}
	all := ctx.Bool("REPORT_ALL_FUNCTIONS")
	switch x := n.(type) {
	case *syntax.FuncCall:
		name, ok := x.Name.(*syntax.Name)
		if !ok {
			return
		}
		r.callback(ctx, x, all)
		if !all && !uqOpcodeFunctions[strings.ToLower(name.Value)] { // D1 (function names are case-insensitive)
			return
		}
		if !uqEligible(ctx, name, false) {
			return
		}
		fqn, fallback := ctx.Names().Function(name.Value, name.Span().Start)
		f := ctx.Index().ResolveFunction(fqn, fallback, ctx.PHP)
		if f == nil || !strings.EqualFold(strings.TrimPrefix(f.FQN, `\`), name.Value) { // D5
			return
		}
		uqReport(ctx, x.Span(), name, name.Value+"(...)")
	case *syntax.ConstFetch:
		if x.Name == nil || !ctx.Bool("REPORT_CONSTANTS") || uqSkippedConstant(x.Name.Value) { // D1, D2
			return
		}
		if !uqEligible(ctx, x.Name, true) {
			return
		}
		fqn, fallback := ctx.Names().Const(x.Name.Value, x.Name.Span().Start)
		c := ctx.Index().Constant(fqn, ctx.PHP)
		if c == nil && fallback != "" {
			c = ctx.Index().Constant(fallback, ctx.PHP)
		}
		if c == nil || strings.TrimPrefix(c.FQN, `\`) != x.Name.Value { // D5
			return
		}
		uqReport(ctx, x.Name.Span(), x.Name, x.Name.Value)
	}
}

func uqReport(ctx *analysis.Context, span syntax.Span, name *syntax.Name, shown string) {
	at := name.Span().Start
	ctx.Report(span, "Write '\\"+shown+"' to allow compile-time binding.", diagnostic.Fix{
		Title: "Qualify with '\\'",
		Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: syntax.Span{Start: at, End: at}, NewText: `\`}}
		},
	})
}

// uqEligible implements D3, D4 and D6.
func uqEligible(ctx *analysis.Context, name *syntax.Name, constant bool) bool {
	if name.NameKind != syntax.NameUnqualified { // D3
		return false
	}
	ns := uqGoverningNamespace(name) // D4
	if ns == nil {
		return false
	}
	for _, st := range ns.Stmts { // D6
		u, ok := st.(*syntax.Use)
		if !ok {
			continue
		}
		for _, it := range u.Items {
			kind := it.Type // the parser gives items the statement's kind
			if (constant && kind != syntax.UseConst) || (!constant && kind != syntax.UseFunction) || it.Name == nil {
				continue
			}
			last := astquery.LastNamePart(it.Name.Value) // the name the import binds: its alias when given
			if it.Alias != nil {
				last = it.Alias.Value
			}
			if last == name.Value || !constant && strings.EqualFold(last, name.Value) {
				return false
			}
		}
	}
	return true
}

// uqGoverningNamespace returns the namespace governing n (D4): its
// enclosing namespace declaration. The parser nests the statements of an
// unbraced `namespace X;` under it, and PHP allows no code outside the
// namespaces of a file that declares one, so a node outside every namespace
// lives in a file without namespace.
func uqGoverningNamespace(n syntax.Node) *syntax.Namespace {
	for p := n.Parent(); p != nil; p = p.Parent() {
		if ns, ok := p.(*syntax.Namespace); ok {
			return ns
		}
	}
	return nil
}

// callback implements Part B (D7–D11).
func (unqualifiedReference) callback(ctx *analysis.Context, call *syntax.FuncCall, all bool) {
	pos, ok := uqCallbackPositions[ctx.GlobalFunctionName(call)] // D7: the global function, any case
	if !ok || astquery.ArgCount(call) < 2 {                      // D7
		return
	}
	if uqGoverningNamespace(call) == nil { // D4 (custos: Part B too)
		return
	}
	arg, ok := call.Args.Args[pos].(*syntax.Arg)
	if !ok || arg.Value == nil {
		return
	}
	content, quote, ok := astquery.QuotedStringRaw(arg.Value)                                       // D8
	if !ok || content == "" || strings.HasPrefix(content, `\`) || strings.Contains(content, "::") { // D9
		return
	}
	if !all && !uqOpcodeFunctions[strings.ToLower(content)] { // D10
		return
	}
	if ctx.Index().Function(content, ctx.PHP) == nil { // D11: a known global function
		return
	}
	span := arg.Value.Span()
	open := span.End - uint32(len(content)) - 1 // position right after the opening quote
	ins := `\`
	if quote == '"' {
		ins = `\\`
	}
	ctx.Report(span, "Write '\\"+content+"' to allow compile-time binding.", diagnostic.Fix{
		Title: "Qualify with '\\'",
		Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: syntax.Span{Start: open, End: open}, NewText: ins}}
		},
	})
}
