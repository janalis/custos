package infer

import (
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/types"
)

// overrideType refines builtin return types that depend on arguments
// (n is a call of the named function name; the parser always builds
// argument lists).
func (e *Env) overrideType(n *syntax.FuncCall, name *syntax.Name) (types.Type, bool) {
	fqn, fb := e.Names.Function(name.Value, name.Span().Start)
	if fb != "" && e.Index.Function(fqn, e.PHP) == nil {
		if e.annotating {
			e.deps = append(e.deps, fqn)
		}
		fqn = fb
	}
	arg := func(i int) syntax.Expr {
		if i < len(n.Args.Args) {
			if a, ok := n.Args.Args[i].(*syntax.Arg); ok && !a.Unpack && a.Name == nil {
				return a.Value
			}
		}
		// Named argument bound to the callee's parameter at position i.
		if f := e.Index.Function(fqn, e.PHP); f != nil && i < len(f.Params) {
			want := strings.TrimPrefix(f.Params[i].Name, "$")
			for _, x := range n.Args.Args {
				if a, ok := x.(*syntax.Arg); ok && a.Name != nil && strings.EqualFold(a.Name.Value, want) {
					return a.Value
				}
			}
		}
		return nil
	}
	switch strings.ToLower(fqn) {
	case "str_replace", "str_ireplace", "preg_replace", "preg_replace_callback", "substr_replace", "preg_filter",
		"preg_replace_callback_array":
		subjectIdx := 2
		switch strings.ToLower(fqn) {
		case "substr_replace":
			subjectIdx = 0
		case "preg_replace_callback_array":
			subjectIdx = 1
		}
		if s := arg(subjectIdx); s != nil {
			st := e.TypeOf(s)
			if st.IsUnknown() {
				return types.Unknown, true // string or array: not known
			}
			if t, ok := replaceResult(st, strings.HasPrefix(strings.ToLower(fqn), "preg_")); ok {
				return t, true
			}
		}
	case "max", "min":
		if t, ok := e.maxMinType(n); ok {
			return t, true
		}
	case "pathinfo":
		// One PATHINFO_* flag returns that part as a string, none the
		// array of parts.
		if len(n.Args.Args) == 1 && arg(0) != nil {
			return types.Array, true
		}
		if c, ok := syntax.UnwrapParens(arg(1)).(*syntax.ConstFetch); ok && strings.HasPrefix(strings.ToUpper(strings.TrimPrefix(c.Name.Value, `\`)), "PATHINFO_") {
			return types.String, true
		}
	case "gettimeofday":
		// gettimeofday(true) is a float, else the array of parts.
		switch {
		case len(n.Args.Args) == 0:
			return types.Array, true
		case constLiteral(arg(0)) == "true":
			return types.Float, true
		case constLiteral(arg(0)) == "false":
			return types.Array, true
		}
	case "var_export", "print_r":
		if t, ok := printReturn(fqn, arg(1), len(n.Args.Args)); ok {
			return t, true
		}
	case "mb_convert_encoding":
		// An array only for an array input.
		if a := arg(0); a != nil {
			st := e.TypeOf(a)
			if st.OnlyOf("string", "int", "float", "bool", "true", "false", "null") {
				return types.Of("string", "false"), true
			}
			if st.IsArrayLike() {
				return types.Of("array", "false"), true
			}
		}
	case "array_rand":
		// One key unless a count other than 1 is asked for.
		if len(n.Args.Args) == 1 {
			return types.Of("int", "string"), true
		}
		if c := arg(1); c != nil {
			if lit, ok := syntax.UnwrapParens(c).(*syntax.Literal); ok && lit.LitKind == syntax.LitInt && lit.Raw == "1" {
				return types.Of("int", "string"), true
			}
		}
	case "current", "reset", "end", "next", "prev", "array_pop", "array_shift":
		if a := arg(0); a != nil {
			at := e.baseType(a)
			el := at.Elem()
			if el.IsUnknown() {
				el = e.shapeElem(at, a)
			} else if v, ok := syntax.UnwrapParens(a).(*syntax.Variable); ok && v.Name != "" && v.Name != "this" {
				el = e.widenVarElem(el, v)
			}
			if !el.IsUnknown() {
				l := strings.ToLower(fqn)
				empty := "false" // the pointer functions return false on an empty array
				if l == "array_pop" || l == "array_shift" {
					empty = "null" // array_pop()/array_shift() return null
				}
				// A provably non-empty array yields an element; current()
				// only while the internal pointer cannot have moved.
				if at.IsArrayLike() && at.IsNonEmptyArray() {
					switch l {
					case "reset", "end", "array_pop", "array_shift":
						return el, true
					case "current":
						if v, ok := syntax.UnwrapParens(a).(*syntax.Variable); ok && v.Name != "" && !e.pointerMoved(syntax.EnclosingVariableScope(v), v.Name) {
							return el, true
						}
					}
				}
				return types.Union(el, types.Of(empty)), true
			}
		}
	case "hrtime":
		// hrtime() / hrtime(false): [seconds, nanoseconds] (or false);
		// hrtime(true): nanoseconds as int (float on 32-bit overflow) or false.
		if len(n.Args.Args) == 0 {
			return types.Of("int[]", "false"), true
		}
		if c, ok := syntax.UnwrapParens(arg(0)).(*syntax.ConstFetch); ok && c.Name != nil {
			switch strings.ToLower(strings.TrimPrefix(c.Name.Value, `\`)) {
			case "false":
				return types.Of("int[]", "false"), true
			case "true":
				return types.Of("int", "float", "false"), true
			}
		}
	case "sscanf":
		// Without output variables the matches are returned (null when
		// the string is empty or input ends early); the int is the count
		// of assigned variables.
		if len(n.Args.Args) == 2 {
			return types.Of("array", "null"), true
		}
	case "parse_url":
		// Without a component the result is the parts array (or false).
		switch len(n.Args.Args) {
		case 1:
			if arg(0) != nil {
				return types.Of("array", "false"), true
			}
		case 2:
			if c, ok := syntax.UnwrapParens(arg(1)).(*syntax.ConstFetch); ok && c != nil {
				if strings.EqualFold(strings.TrimPrefix(c.Name.Value, `\`), "PHP_URL_PORT") {
					return types.Of("int", "null", "false"), true
				}
				if strings.HasPrefix(strings.ToUpper(strings.TrimPrefix(c.Name.Value, `\`)), "PHP_URL_") {
					return types.Of("string", "null", "false"), true
				}
			}
		}
	case "abs":
		if a := arg(0); a != nil {
			if t := e.TypeOf(a); t.OnlyOf("int") || t.OnlyOf("float") {
				return t, true
			}
		}
	case "array_values", "array_reverse", "array_slice", "array_filter", "array_unique":
		if a := arg(0); a != nil {
			if t := e.TypeOf(a); t.IsArrayLike() {
				// Keys are renumbered or dropped: no shape; the result of
				// array_slice()/array_filter() may be empty.
				switch strings.ToLower(fqn) {
				case "array_filter":
					cb := arg(1)
					if cb == nil {
						for _, x := range n.Args.Args {
							a, ok := x.(*syntax.Arg)
							if !ok {
								return t.WithoutShape().WithNonEmpty(false), true
							}
							if a.Unpack {
								cb = a.Value
							}
						}
					}
					return e.arrayFilterType(a, t, cb), true
				case "array_values":
					return t.WithoutShape().WithArrayKey(types.Int), true
				case "array_slice":
					return t.WithoutShape().WithNonEmpty(false).WithArrayKey(types.Of("int", "string")), true
				case "array_reverse":
					return t.WithoutShape().WithArrayKey(types.Of("int", "string")), true
				}
				return t.WithoutShape(), true
			}
		}
	case "array_map":
		if cb, first := arg(0), arg(1); cb != nil && first != nil {
			if t, ok := e.arrayMapType(cb, first, n.Args); ok {
				return t, true
			}
		}
	case "explode":
		// Before PHP 8.0 explode() returns false only for an empty
		// separator (8.0 throws instead).
		if sep, ok := syntax.UnwrapParens(arg(0)).(*syntax.Literal); ok && sep.LitKind == syntax.LitString {
			if v, ok := plainString(sep.Raw); ok && v != "" {
				return types.Of("string[]"), true
			}
		}
	case "array_reduce":
		// The initial value (null when omitted) for an empty array, else
		// the callback's last result.
		if cb := arg(1); cb != nil {
			r := e.callbackReturn(cb)
			init := types.Null
			if x := arg(2); x != nil {
				init = e.TypeOf(x)
			}
			return types.Union(r, init), true // unknown when either is
		}
	case "call_user_func", "call_user_func_array":
		if cb := arg(0); cb != nil {
			if t := e.callbackReturn(cb); !t.IsUnknown() {
				return t, true
			}
		}
	}
	return types.Unknown, false
}

// printReturn types var_export()/print_r(): the output as a string when
// $return (ret, nil when absent or not positional) is literally true, else
// printed: null for var_export, true for print_r. nargs is the number of
// arguments; ok is false when $return is not a literal.
func printReturn(fn string, ret syntax.Expr, nargs int) (types.Type, bool) {
	printed := types.Null
	if strings.EqualFold(fn, "print_r") {
		printed = types.Of("true")
	}
	if ret == nil && nargs < 2 {
		return printed, true
	}
	switch constLiteral(ret) {
	case "true":
		return types.String, true
	case "false":
		return printed, true
	}
	return types.Unknown, false
}

// replaceResult is the result of str_replace()/preg_replace() & co. for a
// subject of type st: a string for scalar or object members (converted),
// an array for array members, plus null for the preg_ functions (a PCRE
// failure). ok is false when st is unknown or holds mixed/iterable.
func replaceResult(st types.Type, preg bool) (types.Type, bool) {
	if st.IsUnknown() || st.HasAny("mixed", "iterable") {
		return types.Unknown, false
	}
	var res []string
	for _, a := range st.Atoms() {
		if a == "array" || strings.HasSuffix(a, "[]") {
			res = append(res, "array")
		} else {
			res = append(res, "string")
		}
	}
	if preg {
		res = append(res, "null")
	}
	return types.Of(res...), true
}

// maxMinType is the type of max()/min(): one of the arguments (two or more
// of them), or an element of the single array argument (false too before
// PHP 8.0 unless the array is known non-empty).
func (e *Env) maxMinType(n *syntax.FuncCall) (types.Type, bool) {
	var ts []types.Type
	for _, x := range n.Args.Args {
		a, ok := x.(*syntax.Arg)
		if !ok || a.Unpack || a.Name != nil {
			return types.Unknown, false
		}
		ts = append(ts, e.TypeOf(a.Value))
	}
	switch len(ts) {
	case 0:
		return types.Unknown, false
	case 1:
		at := ts[0]
		el := iterElem(at)
		if !at.IsArrayLike() || el.IsUnknown() || el.Has("mixed") {
			return types.Unknown, false
		}
		if e.PHP < phpversion.PHP80 && !at.IsNonEmptyArray() {
			el = types.Union(el, types.Of("false"))
		}
		return el.WithoutArrayInfo(), true
	}
	u := types.Union(ts...)
	if u.IsUnknown() || u.Has("mixed") {
		return types.Unknown, false
	}
	return u.WithoutArrayInfo(), true
}
