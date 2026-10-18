package unescapedhtmloutput

import (
	"maps"
	"slices"
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/flow"
)

// maxOrigins bounds the chain of assignments followed for one output value.
const maxOrigins = 8

// tainted narrows the flow answer for an emitted value. The flow engine joins
// the taint of every operand, so values that only select between trusted
// results, or whose string conversion is a number or boolean, are cleared
// here. It never turns an untainted value into a tainted one.
func tainted(ctx *analysis.Context, expr syntax.Expr, depth int, context flow.Context) bool {
	expr = syntax.UnwrapParens(expr)
	if !ctx.Flow().Tainted(expr, context) {
		return false
	}
	if depth == maxOrigins {
		return true
	}
	switch x := expr.(type) {
	case *syntax.Variable:
		if origin, known := ctx.Flow().Resolve(x); known && origin != expr && !aliased(ctx, x) {
			return tainted(ctx, origin, depth+1, context)
		}
	case *syntax.Ternary:
		return tainted(ctx, x.Then, depth+1, context) || tainted(ctx, x.Else, depth+1, context)
	case *syntax.Match:
		// Only arm results are emitted; the subject and arm conditions merely
		// select one of them.
		for _, arm := range x.Arms {
			if tainted(ctx, arm.Body, depth+1, context) {
				return true
			}
		}
		return false
	case *syntax.Binary:
		// Arithmetic yields a number or throws. Array union keeps request
		// strings as elements, but the emitted conversion is then "Array".
		switch x.Op.Kind {
		case syntax.TPlus, syntax.TMinus, syntax.TMul, syntax.TDiv, syntax.TMod, syntax.TPow:
			return false
		}
	case *syntax.Unary:
		// Arithmetic negation and identity yield a number or throw.
		return x.Op.Kind != syntax.TMinus && x.Op.Kind != syntax.TPlus
	case *syntax.FuncCall:
		return !scalarCall(ctx, x)
	}
	return true
}

// aliased reports whether the scope of v can rebind it without an assignment
// the resolved value accounts for: a reference to it (alias, by-reference
// capture or argument), a catch binding or static declaration of the same
// name, extract(), an assignment through a variable variable or the global
// symbol table.
func aliased(ctx *analysis.Context, v *syntax.Variable) bool {
	found := false
	named := func(e syntax.Expr) bool { return rootVariable(e) == v.Name }
	scope := syntax.EnclosingVariableScope(v)
	local := func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.Assign:
			target, _ := syntax.UnwrapParens(x.Var).(*syntax.Variable)
			found = found || target != nil && target.NameExpr != nil || x.ByRef && (named(x.Var) || named(x.Value))
		case *syntax.Foreach:
			found = found || x.ByRef && named(x.Expr)
		case *syntax.Closure:
			for _, use := range x.Uses {
				found = found || use.ByRef && use.Var.Name == v.Name
			}
		case *syntax.Catch:
			found = found || x.Var != nil && x.Var.Name == v.Name
		case *syntax.StaticVar:
			found = found || x.Var.Name == v.Name
		case *syntax.FuncCall:
			found = found || semanticquery.NativeBuiltinName(ctx, x) == "extract"
			for _, target := range referenceArguments(ctx, x) {
				found = found || named(target.Argument.Value)
			}
		}
		return !found && (n == scope || !syntax.IsVariableScope(n))
	}
	// The global symbol table is shared with every function of the file.
	shared := func(n syntax.Node) bool {
		if x, ok := n.(*syntax.Variable); ok {
			_, global := x.Parent().(*syntax.Global)
			key, known := elementKey(ctx, x)
			found = found || x.Name == "GLOBALS" && (!known || key == v.Name) || global && x.Name == v.Name
		}
		return !found
	}
	if scope != nil {
		syntax.Inspect(scope, local)
		syntax.Inspect(scope, shared)
	} else {
		syntax.InspectFile(ctx.File, local)
		syntax.InspectFile(ctx.File, shared)
	}
	return found
}

// scalarCall proves a resolved sprintf builds its text from a literal format
// with numeric conversions only. Builtins that return a number or a boolean
// carry no taint in the flow analysis and never reach this point.
func scalarCall(ctx *analysis.Context, call *syntax.FuncCall) bool {
	if semanticquery.NativeBuiltinName(ctx, call) != "sprintf" {
		return false
	}
	format, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 0, "format"))
	return known && numericFormat(format)
}

// numericFormat accepts only literal percent signs and numeric conversions.
// Custom padding characters and `%c`, which can emit any byte, are rejected.
func numericFormat(format string) bool {
	for i := 0; i < len(format); i++ {
		if format[i] != '%' {
			continue
		}
		i++
		for i < len(format) && strings.IndexByte("0123456789$+-. ", format[i]) >= 0 {
			i++
		}
		if i == len(format) || strings.IndexByte("%bdeEfFgGouxX", format[i]) < 0 {
			return false
		}
	}
	return true
}

// requestRewritten reports whether a request superglobal feeding the input is
// written, unset or passed by reference earlier in the same scope. Such a
// write may replace the request text, so the source-to-use path is unproven.
// A write to another literal key than the ones read, or one that stores a
// value still holding request text, replaces nothing the output depends on.
// A reference bound to the superglobal replaces nothing either: writes to the
// bound name are then read as writes to the superglobal.
// The writes are read in source order: one that stores request text into the
// whole table, into the element a replacement emptied, or into an element no
// literal names (a computed or nested key, a reference to the table, an entry
// of `$GLOBALS`), fills it again. So does code the scope does not show: a call
// or a construction not known to spare the request tables (see
// flow.Env.Spares) may store request text in any of them. In a loop
// around the output, what fills a table after the output does so before its
// next turn.
func requestRewritten(ctx *analysis.Context, at syntax.Node, input syntax.Expr, context flow.Context) bool {
	names := map[string]bool{"GLOBALS": true}
	spans := map[syntax.Span]bool{}
	for _, source := range ctx.Flow().Value(input).Sources {
		if source.Parameter < 0 {
			names[source.Kind] = true
			spans[source.Span] = true
		}
	}
	// A write's escaping contract applies to a later direct read, but not
	// to an output transform that can decode that text again.
	escapingApplies := directRequestRead(ctx, input, 0) || context == flow.HTML && preservingRequestRead(ctx, input, 0)
	scope := syntax.EnclosingVariableScope(at)
	limit := at.Span().Start
	// whole lists the superglobals read as a whole or through an unknown key;
	// keys lists the literal elements read, as name and key.
	whole, keys := map[string]bool{"GLOBALS": true}, map[[2]string]bool{}
	located := 0
	type write struct {
		target  syntax.Expr
		replace bool // false when the stored value still holds request text
		refill  bool // the write stores request text
		any     bool // code that may store request text in any table
	}
	var writes []write
	// again is where the outermost loop around the output ends: what is
	// written up to there may run before the output does once more.
	again := limit
	for p := at.Parent(); p != nil && p != scope; p = p.Parent() {
		switch p.(type) {
		case *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach:
			again = p.Span().End
		}
	}
	late := false
	add := func(w write) {
		if late && !w.refill {
			// A replacement made after the output says nothing of the
			// first turn.
			return
		}
		w.replace = w.replace && !late
		writes = append(writes, w)
	}
	// called is the value of the assignment whose calls are already counted:
	// they run before the assignment stores its result.
	var called syntax.Span
	foreign := func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.MethodCall, *syntax.StaticCall, *syntax.New:
			return !ctx.Flow().Spares(n)
		case *syntax.FuncCall:
			// What a PHP function is given by reference is read below.
			return semanticquery.NativeBuiltinName(ctx, x) == "" && !ctx.Flow().Spares(n)
		}
		return false
	}
	// alias maps a variable bound by reference to the superglobal it shares,
	// or to the element of it a literal key names.
	type shared struct {
		table, key string
		// whole tells that the name stands for the table itself, and named
		// that it stands for the element key names.
		whole, named bool
	}
	alias := map[string]shared{}
	bind := func(from, to syntax.Expr) {
		name, target := rootVariable(from), shared{table: rootVariable(to), whole: true}
		if fetch, ok := syntax.UnwrapParens(to).(*syntax.ArrayDimFetch); ok {
			// Only an element directly in the table has a key to name.
			target.whole = false
			if v, direct := syntax.UnwrapParens(fetch.Var).(*syntax.Variable); direct {
				target.key, target.named = elementKey(ctx, v)
			}
		}
		if known, ok := alias[rootVariable(to)]; ok {
			target = known
		}
		if name != "" && !names[name] && names[target.table] {
			alias[name] = target
		}
	}
	visit := func(n syntax.Node) bool {
		if n != scope && syntax.IsVariableScope(n) {
			return false
		}
		if v, ok := n.(*syntax.Variable); ok && spans[v.Span()] {
			located++
			if key, known := elementKey(ctx, v); known {
				keys[[2]string{v.Name, key}] = true
			} else {
				whole[v.Name] = true
			}
		}
		if late = n.Span().Start >= limit; n.Span().Start >= again {
			return true
		}
		if foreign(n) && !called.Contains(n.Span()) {
			add(write{any: true, refill: true})
		}
		switch x := n.(type) {
		case *syntax.Assign:
			if x.Value != nil && !called.Contains(x.Span()) {
				called = x.Value.Span()
				syntax.Inspect(x.Value, func(child syntax.Node) bool {
					if foreign(child) {
						add(write{any: true, refill: true})
						return false
					}
					return !syntax.IsVariableScope(child)
				})
			}
			clean := !ctx.Flow().Tainted(x.Value, context) && (escapingApplies || !requestSource(ctx.Flow().Value(x.Value))) || context == flow.HTML && callbackResult(ctx, x.Value)
			for _, target := range slots(x.Var) {
				add(write{target: target, replace: x.Op.Kind == syntax.TEqual && clean, refill: !clean})
			}
			if x.ByRef {
				bind(x.Var, x.Value)
				bind(x.Value, x.Var)
			}
		case *syntax.IncDec:
			add(write{target: x.Var, replace: true})
		case *syntax.Unset:
			for _, target := range x.Vars {
				add(write{target: target, replace: true})
			}
		case *syntax.Foreach:
			// The targets receive the keys and elements of what is iterated.
			fed := ctx.Flow().Tainted(x.Expr, context) || !escapingApplies && requestSource(ctx.Flow().Value(x.Expr))
			for _, target := range append(slots(x.Key), slots(x.Value)...) {
				add(write{target: target, replace: !fed, refill: fed})
			}
			if x.ByRef {
				add(write{target: x.Expr, replace: true})
			}
		case *syntax.FuncCall:
			targets := referenceArguments(ctx, x)
			fed := len(targets) > 0 && fedByRequest(ctx, x, targets, context, escapingApplies)
			for _, reference := range targets {
				arg := reference.Argument
				replace, refill := !fed && reference.Known && escapingApplies, fed || !reference.Known
				if semanticquery.NativeBuiltinName(ctx, x) == "" {
					value, known := ctx.Flow().ReferenceResult(x, arg)
					replace = known && (escapingApplies && value.Safe&context != 0 || !requestSource(value) || context == flow.HTML && callbackResult(ctx, value.Expr))
					refill = !replace
				}
				add(write{target: arg.Value, replace: replace, refill: refill})
			}
		}
		return true
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	// A source read in another scope (through a callee) has no known key.
	unknown := located != len(spans)
	// emptied holds the elements a replacement left without request text, and
	// cleared the tables replaced as a whole or through an element no literal
	// names; filled then lists the elements of such a table written since.
	emptied, filled := map[[2]string]bool{}, map[[2]string]bool{}
	cleared := map[string]bool{}
	for _, w := range writes {
		name := rootVariable(w.target)
		through, bound := alias[name]
		if bound {
			name = through.table
		}
		if w.any || w.refill && name == "GLOBALS" {
			// Request text may now be in any element of any table.
			clear(emptied)
			clear(cleared)
			continue
		}
		if !names[name] {
			continue
		}
		root, element := syntax.UnwrapParens(w.target).(*syntax.ArrayDimFetch)
		key, known := through.key, through.named
		if element && (!bound || through.whole) {
			if v, direct := syntax.UnwrapParens(root.Var).(*syntax.Variable); direct {
				key, known = elementKey(ctx, v)
			}
		}
		slot := [2]string{name, key}
		switch {
		case w.replace && !known:
			cleared[name] = true
			maps.DeleteFunc(filled, func(s [2]string, _ bool) bool { return s[0] == name })
		case w.replace:
			if unknown || whole[name] || keys[slot] {
				emptied[slot] = true
			}
			delete(filled, slot)
		case w.refill && known:
			// Request text stored into one element reaches that element.
			delete(emptied, slot)
			filled[slot] = true
		case w.refill:
			// Stored into the whole table, it reaches every element; stored
			// into an element no literal names, it may reach any of them.
			delete(cleared, name)
			maps.DeleteFunc(emptied, func(s [2]string, _ bool) bool { return s[0] == name })
		}
	}
	for name := range cleared {
		if unknown || whole[name] {
			return true
		}
		for slot := range keys {
			if slot[0] == name && !filled[slot] {
				return true
			}
		}
	}
	return len(emptied) > 0
}

// elementKey returns the literal key of the element access whose base is v.
func elementKey(ctx *analysis.Context, v *syntax.Variable) (string, bool) {
	fetch, ok := v.Parent().(*syntax.ArrayDimFetch)
	if !ok || fetch.Var != syntax.Expr(v) || fetch.Dim == nil {
		return "", false
	}
	if key, known := semanticquery.NativeString(ctx, fetch.Dim); known {
		return key, true
	}
	if key, known := semanticquery.NativeInt(ctx, fetch.Dim); known {
		return strconv.FormatInt(key, 10), true
	}
	return "", false
}

// pointerOnly lists builtins that take an array by reference without
// changing its elements: they read it or move its internal pointer.
var pointerOnly = map[string]bool{
	"extract": true, "current": true, "key": true, "end": true, "reset": true, "next": true, "prev": true, "each": true,
}

// referenceArguments returns arguments that may bind to by-reference
// parameters of a resolved callee that may change them.
func referenceArguments(ctx *analysis.Context, call *syntax.FuncCall) []semanticquery.ReferenceArgument {
	callee, ok := semanticquery.ResolveCallee(ctx, call)
	if !ok || call.Args == nil || pointerOnly[semanticquery.NativeBuiltinName(ctx, call)] {
		return nil
	}
	return semanticquery.ReferenceArguments(call.Args, callee.Params, true)
}

// callbackResult reports the result of a resolved function that applies a
// callback: the taint cannot tell whether `array_map('htmlspecialchars', …)`
// still holds request text.
func callbackResult(ctx *analysis.Context, value syntax.Expr) bool {
	call, ok := syntax.UnwrapParens(value).(*syntax.FuncCall)
	if !ok {
		return false
	}
	callee, _ := semanticquery.ResolveCallee(ctx, call)
	for _, param := range callee.Params {
		if strings.Contains(param.Type, "callable") {
			return true
		}
	}
	return false
}

// fedByRequest reports a builtin that receives request text in an argument
// it does not write: what it stores in the referenced ones may derive from it
// (`parse_str($_SERVER['QUERY_STRING'], $_GET)`).
func fedByRequest(ctx *analysis.Context, call *syntax.FuncCall, written []semanticquery.ReferenceArgument, context flow.Context, escapingApplies bool) bool {
	if semanticquery.NativeBuiltinName(ctx, call) == "" {
		return false
	}
	for _, node := range call.Args.Args {
		if arg, ok := node.(*syntax.Arg); ok && !slices.ContainsFunc(written, func(reference semanticquery.ReferenceArgument) bool { return reference.Argument == arg }) && (ctx.Flow().Tainted(arg.Value, context) || !escapingApplies && requestSource(ctx.Flow().Value(arg.Value))) {
			return true
		}
	}
	return false
}

// slots returns the targets a write target stands for: itself, or each slot
// of a destructuring.
func slots(target syntax.Expr) []syntax.Expr {
	var items []*syntax.ArrayItem
	switch x := syntax.UnwrapParens(target).(type) {
	case nil:
		return nil
	case *syntax.Array:
		items = x.Items
	case *syntax.List:
		items = x.Items
	default:
		return []syntax.Expr{target}
	}
	var out []syntax.Expr
	for _, item := range items {
		if item != nil {
			out = append(out, slots(item.Value)...)
		}
	}
	return out
}

// rootVariable names the variable an element or whole-variable target writes.
func rootVariable(target syntax.Expr) string {
	for {
		switch x := syntax.UnwrapParens(target).(type) {
		case *syntax.ArrayDimFetch:
			target = x.Var
			continue
		case *syntax.Variable:
			return x.Name
		}
		return ""
	}
}

// directRequestRead distinguishes a read (including a reaching local copy)
// from a transform that can invalidate a prior write's escaping contract.
func directRequestRead(ctx *analysis.Context, input syntax.Expr, depth int) bool {
	if depth == maxOrigins {
		return false
	}
	switch x := syntax.UnwrapParens(input).(type) {
	case *syntax.ArrayDimFetch:
		return directRequestRead(ctx, x.Var, depth+1)
	case *syntax.Variable:
		switch x.Name {
		case "_GET", "_POST", "_REQUEST", "_COOKIE", "_FILES", "_SERVER", "GLOBALS":
			return x.NameExpr == nil
		}
		if origin, known := ctx.Flow().Resolve(x); known && origin != x && !aliased(ctx, x) {
			return directRequestRead(ctx, origin, depth+1)
		}
	}
	return false
}

func requestSource(value flow.Value) bool {
	return slices.ContainsFunc(value.Sources, func(source flow.Source) bool { return source.Parameter < 0 })
}

// preservingRequestRead recognizes the narrow native transforms that cannot
// expose markup from an earlier escaped request write.
func preservingRequestRead(ctx *analysis.Context, input syntax.Expr, depth int) bool {
	if depth == maxOrigins {
		return false
	}
	if directRequestRead(ctx, input, depth) {
		return true
	}
	call, ok := syntax.UnwrapParens(input).(*syntax.FuncCall)
	if !ok {
		return false
	}
	switch semanticquery.NativeBuiltinName(ctx, call) {
	case "trim", "ltrim", "rtrim", "strtolower", "strtoupper", "ucfirst", "lcfirst", "ucwords", "str_repeat", "strrev", "substr":
		arg, state := semanticquery.CallArgumentBinding(call.Args, 0, "string")
		return state == semanticquery.ArgumentBound && preservingRequestRead(ctx, arg, depth+1)
	}
	return false
}
