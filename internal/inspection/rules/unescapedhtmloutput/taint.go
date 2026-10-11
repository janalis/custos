package unescapedhtmloutput

import (
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
func tainted(ctx *analysis.Context, expr syntax.Expr, depth int) bool {
	expr = syntax.UnwrapParens(expr)
	if !ctx.Flow().Tainted(expr, flow.HTML) {
		return false
	}
	if depth == maxOrigins {
		return true
	}
	switch x := expr.(type) {
	case *syntax.Variable:
		if origin, known := ctx.Flow().Resolve(x); known && origin != expr && !aliased(ctx, x) {
			return tainted(ctx, origin, depth+1)
		}
	case *syntax.Ternary:
		return tainted(ctx, x.Then, depth+1) || tainted(ctx, x.Else, depth+1)
	case *syntax.Match:
		// Only arm results are emitted; the subject and arm conditions merely
		// select one of them.
		for _, arm := range x.Arms {
			if tainted(ctx, arm.Body, depth+1) {
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
				found = found || named(target)
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

// scalarCall proves a resolved builtin returns only an integer, a boolean or
// text built from a literal format with numeric conversions.
func scalarCall(ctx *analysis.Context, call *syntax.FuncCall) bool {
	name := semanticquery.NativeBuiltinName(ctx, call)
	if strings.HasPrefix(name, "is_") || strings.HasPrefix(name, "ctype_") {
		return true
	}
	switch name {
	case "strlen", "mb_strlen", "count", "sizeof", "crc32", "ord", "strcmp", "strcasecmp", "strncmp", "strncasecmp",
		"strpos", "stripos", "strrpos", "strripos", "mb_strpos", "levenshtein", "in_array", "array_key_exists", "strtotime":
		return true
	case "sprintf":
		format, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 0, "format"))
		return known && numericFormat(format)
	}
	return false
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
func requestRewritten(ctx *analysis.Context, at syntax.Node, input syntax.Expr) bool {
	names := map[string]bool{"GLOBALS": true}
	spans := map[syntax.Span]bool{}
	for _, source := range ctx.Flow().Value(input).Sources {
		if source.Parameter < 0 {
			names[source.Kind] = true
			spans[source.Span] = true
		}
	}
	scope := syntax.EnclosingVariableScope(at)
	limit := at.Span().Start
	// whole lists the superglobals read as a whole or through an unknown key;
	// keys lists the literal elements read, as name and key.
	whole, keys := map[string]bool{"GLOBALS": true}, map[[2]string]bool{}
	located := 0
	type write struct {
		target  syntax.Expr
		replace bool // false when the stored value still holds request text
	}
	var writes []write
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
		if n.Span().Start >= limit {
			return true
		}
		switch x := n.(type) {
		case *syntax.Assign:
			writes = append(writes, write{x.Var, x.Op.Kind == syntax.TEqual && !ctx.Flow().Tainted(x.Value, flow.HTML)})
			if x.ByRef {
				writes = append(writes, write{x.Var, true}, write{x.Value, true})
			}
		case *syntax.IncDec:
			writes = append(writes, write{x.Var, true})
		case *syntax.Unset:
			for _, target := range x.Vars {
				writes = append(writes, write{target, true})
			}
		case *syntax.Foreach:
			writes = append(writes, write{x.Key, true}, write{x.Value, true})
			if x.ByRef {
				writes = append(writes, write{x.Expr, true})
			}
		case *syntax.FuncCall:
			for _, target := range referenceArguments(ctx, x) {
				writes = append(writes, write{target, true})
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
	for _, w := range writes {
		name := rootVariable(w.target)
		if !names[name] {
			continue
		}
		root, element := syntax.UnwrapParens(w.target).(*syntax.ArrayDimFetch)
		key, known := "", false
		if element {
			if v, direct := syntax.UnwrapParens(root.Var).(*syntax.Variable); direct {
				key, known = elementKey(ctx, v)
			}
		}
		if !known {
			return true
		}
		if w.replace && (unknown || whole[name] || keys[[2]string{name, key}]) {
			return true
		}
	}
	return false
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

// referenceArguments returns the positional arguments bound to by-reference
// parameters of a resolved callee that may change them.
func referenceArguments(ctx *analysis.Context, call *syntax.FuncCall) []syntax.Expr {
	callee, ok := semanticquery.ResolveCallee(ctx, call)
	if !ok || call.Args == nil || pointerOnly[semanticquery.NativeBuiltinName(ctx, call)] {
		return nil
	}
	var out []syntax.Expr
	for i, node := range call.Args.Args {
		if arg, ok := node.(*syntax.Arg); ok && i < len(callee.Params) && callee.Params[i].ByRef {
			out = append(out, arg.Value)
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
