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
		if uqRuntimeDefined(ctx, x.Name) { // D5: a define() the index cannot name
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
	value := uqArgument(call, pos, "callback")                                                      // D8: the argument bound to the callback
	content, quote, ok := astquery.QuotedStringRaw(value)                                           // D8
	if !ok || content == "" || strings.HasPrefix(content, `\`) || strings.Contains(content, "::") { // D9
		return
	}
	if !all && !uqOpcodeFunctions[strings.ToLower(content)] { // D10
		return
	}
	if ctx.Index().Function(content, ctx.PHP) == nil { // D11: a known global function
		return
	}
	span := value.Span()
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

// uqArgument returns the value bound to the parameter at pos, named param:
// a named argument binds by its name whatever its position. Nil when the
// parameter is not passed or an unpacked argument hides the positions.
func uqArgument(call *syntax.FuncCall, pos int, param string) syntax.Expr {
	i := 0
	for _, n := range call.Args.Args {
		arg, ok := n.(*syntax.Arg)
		if !ok || arg.Unpack {
			return nil
		}
		if arg.Name != nil {
			if arg.Name.Value == param {
				return arg.Value
			}
			continue
		}
		if i == pos {
			return arg.Value
		}
		i++
	}
	return nil
}

// uqRuntimeDefined reports whether the file holds a define() call that
// declares name in its governing namespace under a spelling the index does
// not record (D5): a name built by concatenation (`__NAMESPACE__ . '\X'`)
// or written with an escaped backslash.
func uqRuntimeDefined(ctx *analysis.Context, name *syntax.Name) bool {
	defined := ctx.Memo("defines", func() any {
		set := map[string]bool{}
		syntax.InspectFile(ctx.File, func(n syntax.Node) bool {
			call, ok := n.(*syntax.FuncCall)
			if !ok || call.Args == nil {
				return true
			}
			fn, ok := call.Name.(*syntax.Name)
			if !ok || !strings.EqualFold(astquery.LastNamePart(fn.Value), "define") {
				return true
			}
			full, exact := uqFoldName(ctx, uqArgument(call, 0, "constant_name"))
			if exact {
				full = strings.TrimPrefix(full, `\`)
			}
			if cut := strings.LastIndexByte(full, '\\'); cut >= 0 {
				if !exact { // unknown prefix: any namespace may be meant
					set[full[cut:]] = true
				}
				set[strings.ToLower(full[:cut])+full[cut:]] = true
			} else if !exact && full != "" { // `$ns . 'X'`: the prefix may be a namespace
				set[`\`+full] = true
			}
			return true
		})
		return set
	}).(map[string]bool)
	ns := ""
	if decl := uqGoverningNamespace(name); decl.Name != nil {
		ns = strings.ToLower(decl.Name.Value)
	}
	return defined[ns+`\`+name.Value] || defined[`\`+name.Value]
}

// uqFoldName evaluates the name argument of a define() call: string
// literals, `__NAMESPACE__` and their concatenations. exact is false when
// only the trailing literal of a concatenation is known; the text is then
// that tail.
func uqFoldName(ctx *analysis.Context, e syntax.Expr) (text string, exact bool) {
	switch x := syntax.UnwrapParens(e).(type) {
	case *syntax.Binary:
		if x.Op.Kind != syntax.TDot {
			return "", false
		}
		right, exact := uqFoldName(ctx, x.Right)
		if !exact {
			return right, false
		}
		left, exact := uqFoldName(ctx, x.Left)
		return left + right, exact
	case *syntax.MagicConst:
		if !strings.EqualFold(ctx.Text(x), "__NAMESPACE__") {
			return "", false
		}
		if ns := uqGoverningNamespace(x); ns != nil && ns.Name != nil {
			return ns.Name.Value, true
		}
		return "", true // the global namespace
	case *syntax.Literal:
		if raw, _, ok := astquery.QuotedStringRaw(x); ok {
			return strings.ReplaceAll(raw, `\\`, `\`), true
		}
	}
	return "", false
}
