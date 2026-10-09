package infer

import (
	"strings"

	"custos/internal/index"
	"custos/internal/stubs"
	"custos/internal/syntax"
	"custos/internal/types"
)

// Out parameters: a variable passed to a by-reference parameter that
// documents the type it leaves (`@param-out T $x`, phpstan-/psalm-
// variants), or to an output parameter of a builtin (preg_match()'s
// $matches, exec()'s $output, …), is defined by the call: after it, the
// variable has that type. The definition hides earlier ones like an
// assignment when the call always runs with its statement (see
// unconditionalBlock). Callees are resolved without type inference (named
// functions, `Name::`/`self::`/`parent::`/`static::` calls, `$this->`
// calls and `new Name`), since definitions are collected before any
// variable is typed.

// outKind describes the value a builtin stores in an output parameter.
type outKind uint8

const (
	outInt         outKind = iota + 1
	outFloat               // float
	outString              // string
	outArray               // array of unknown elements
	outMatches             // preg_match: string[] without flags, else array
	outMatchesAll          // preg_match_all: string[][] without flags, else array
	outAppendLines         // exec: string lines appended to an array
)

// builtinOut lists, per builtin function, the output parameters by
// position and the value they hold after the call.
var builtinOut = map[string]map[int]outKind{
	"preg_match":                  {2: outMatches},
	"preg_match_all":              {2: outMatchesAll},
	"preg_replace":                {4: outInt},
	"preg_replace_callback":       {4: outInt},
	"preg_replace_callback_array": {3: outInt},
	"preg_filter":                 {4: outInt},
	"str_replace":                 {3: outInt},
	"str_ireplace":                {3: outInt},
	"parse_str":                   {1: outArray},
	"mb_parse_str":                {1: outArray},
	"exec":                        {1: outAppendLines, 2: outInt},
	"system":                      {1: outInt},
	"passthru":                    {1: outInt},
	"similar_text":                {2: outFloat},
	"fsockopen":                   {2: outInt, 3: outString},
	"pfsockopen":                  {2: outInt, 3: outString},
	"stream_socket_client":        {1: outInt, 2: outString},
	"stream_socket_server":        {1: outInt, 2: outString},
	"headers_sent":                {0: outString, 1: outInt},
	"getimagesize":                {1: outArray},
	"getimagesizefromstring":      {1: outArray},
	"is_callable":                 {2: outString},
	"openssl_sign":                {1: outString},
	"openssl_public_encrypt":      {1: outString},
	"openssl_private_encrypt":     {1: outString},
	"openssl_public_decrypt":      {1: outString},
	"openssl_private_decrypt":     {1: outString},
	"getopt":                      {2: outInt},
	"proc_open":                   {2: outArray},
	"flock":                       {2: outInt},
}

// collectOutArgs records the variables call (a function, method or static
// call, or `new`) defines through out parameters.
func (e *Env) collectOutArgs(call syntax.Expr, sv *scopeVars) {
	var al *syntax.ArgList
	switch c := call.(type) {
	case *syntax.FuncCall:
		al = c.Args
	case *syntax.MethodCall:
		al = c.Args
	case *syntax.StaticCall:
		al = c.Args
	case *syntax.New:
		al = c.Args
	}
	if al == nil || !hasVarArg(al) {
		return
	}
	params, builtin := e.calleeParams(call)
	if params == nil {
		return
	}
	var kill syntax.Span
	killSet := false // unconditionalBlock is computed once per call
	for i, x := range al.Args {
		a, ok := x.(*syntax.Arg)
		if !ok || a.Unpack {
			continue
		}
		v, ok := a.Value.(*syntax.Variable)
		if !ok || v.Name == "" || v.Name == "this" {
			continue
		}
		p, ok := paramFor(params, a, i)
		if !ok || !p.ByRef {
			continue
		}
		var typ func() types.Type
		if builtin != nil {
			if kind := builtin[paramIndex(params, p)]; kind != 0 {
				typ = e.builtinOutType(kind, al, v)
			}
		} else if p.Out != "" && !e.native {
			out := p.Out
			typ = func() types.Type {
				if t := types.FromDoc(out, nil); !t.Has("mixed") {
					return t
				}
				return types.Unknown
			}
		}
		if typ == nil {
			continue
		}
		if !killSet {
			kill, killSet = unconditionalBlock(call), true
			if arrow := e.arrowDefinitionSpan(call); arrow.Len() > 0 {
				kill = arrow
			}
		}
		sv.defs[v.Name] = append(sv.defs[v.Name], varDef{pos: call.Span().Start, end: call.Span().End, kill: kill, typ: typ})
	}
}

// hasVarArg reports whether an argument list passes a plain variable.
func hasVarArg(al *syntax.ArgList) bool {
	for _, x := range al.Args {
		if a, ok := x.(*syntax.Arg); ok {
			if _, ok := a.Value.(*syntax.Variable); ok {
				return true
			}
		}
	}
	return false
}

// paramIndex is the position of p (one of params) in params.
func paramIndex(params []index.Param, p index.Param) int {
	i := 0
	for i < len(params)-1 && params[i].Name != p.Name {
		i++
	}
	return i
}

// calleeParams resolves the parameters of call without type inference;
// builtin lists the output parameters when the callee is a builtin
// function known to builtinOut (nil otherwise).
func (e *Env) calleeParams(call syntax.Expr) (params []index.Param, builtin map[int]outKind) {
	var m *index.Method
	switch c := call.(type) {
	case *syntax.FuncCall:
		f := e.ResolveFunction(c)
		if f == nil {
			return nil, nil
		}
		if stubs.Index().Function(f.FQN, e.PHP) == f {
			if b, ok := builtinOut[strings.ToLower(f.FQN)]; ok {
				return f.Params, b
			}
		}
		return f.Params, nil
	case *syntax.MethodCall:
		id, ok := c.Name.(*syntax.Identifier)
		recv, isVar := syntax.UnwrapParens(c.Var).(*syntax.Variable)
		if !ok || !isVar || recv.Name != "this" || recv.NameExpr != nil {
			return nil, nil
		}
		if cls := e.selfClass(syntax.EnclosingClass(c)); cls != "" {
			m = e.Index.FindMethod(cls, id.Value, e.PHP)
		}
	case *syntax.StaticCall:
		id, ok := c.Name.(*syntax.Identifier)
		nm, named := c.Class.(*syntax.Name)
		if !ok || !named {
			return nil, nil
		}
		if cls := e.classRef(nm); cls != "" {
			m = e.Index.FindMethod(cls, id.Value, e.PHP)
		}
	case *syntax.New:
		if nm, ok := c.Class.(*syntax.Name); ok {
			if cls := e.classRef(nm); cls != "" {
				m = e.Index.FindMethod(cls, "__construct", e.PHP)
			}
		}
	}
	if m == nil {
		return nil, nil
	}
	return m.Params, nil
}

// builtinOutType is the type a builtin's output parameter of kind holds
// after the call (al: its arguments; v: the variable passed).
func (e *Env) builtinOutType(kind outKind, al *syntax.ArgList, v *syntax.Variable) func() types.Type {
	switch kind {
	case outInt:
		return func() types.Type { return types.Int }
	case outFloat:
		return func() types.Type { return types.Float }
	case outString:
		return func() types.Type { return types.String }
	case outMatches, outMatchesAll:
		// Unmatched groups are '' (trailing ones absent) without flags;
		// PREG_UNMATCHED_AS_NULL makes them null. Other flags
		// (PREG_OFFSET_CAPTURE, PREG_SET_ORDER…) change the entries'
		// shape: plain array.
		el := "string"
		if len(al.Args) > 3 {
			switch matchFlags(al.Args[3]) {
			case "0":
			case "preg_unmatched_as_null":
				el = "(string|null)"
			default:
				el = ""
			}
		}
		return func() types.Type {
			switch {
			case el == "":
				return types.Array
			case kind == outMatches:
				return types.FromDoc(el+"[]", nil)
			}
			return types.FromDoc(el+"[][]", nil)
		}
	case outAppendLines:
		// exec() appends to an array (a non-array becomes one): the lines
		// join the elements it already had.
		return func() types.Type {
			prior := e.TypeOf(v)
			if prior.IsArrayLike() && !prior.Has("array") {
				return arrayOf(types.Union(iterElem(prior), types.String))
			}
			if prior.IsArrayLike() {
				return types.Array
			}
			return types.Of("string[]")
		}
	}
	return func() types.Type { return types.Array } // outArray
}

// unconditionalBlock returns the span of the block in which call runs
// whenever the statement holding it does (the statement sits directly in
// that block): the call is not inside the right operand of `&&`, `||`,
// `??`, a ternary branch, a closure, a match arm…, and the statement is an
// expression statement, a return/echo, or the condition of an `if`,
// `while` or `switch` (evaluated at least once). Zero span: none.
func unconditionalBlock(call syntax.Expr) syntax.Span {
	var child syntax.Node = call
	for p := call.Parent(); ; child, p = p, p.Parent() {
		switch n := p.(type) {
		case *syntax.Paren, *syntax.Unary, *syntax.Arg, *syntax.ArgList, *syntax.Assign,
			*syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall, *syntax.New, *syntax.ArrayDimFetch:
			continue
		case *syntax.Binary:
			switch n.Op.Kind {
			case syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr, syntax.TCoalesce:
				if child != syntax.Node(n.Left) {
					return syntax.Span{}
				}
			}
			continue
		case *syntax.Ternary:
			if child != syntax.Node(n.Cond) {
				return syntax.Span{}
			}
			continue
		case *syntax.ExprStmt, *syntax.Return, *syntax.Echo, *syntax.If, *syntax.While, *syntax.Switch:
			// Reached from their condition: a call in a body stops at its
			// own statement, and elseif/else/case nodes are not listed.
		default:
			return syntax.Span{} // nil included (no enclosing statement)
		}
		if blk, ok := p.Parent().(*syntax.Block); ok {
			return blk.Span()
		}
		return syntax.Span{}
	}
}

// matchFlags names the flags argument of preg_match(): "0" for a literal 0,
// the lower-case constant name for a single constant, "" otherwise.
func matchFlags(x syntax.Expr) string {
	a, ok := x.(*syntax.Arg)
	if !ok || a.Name != nil || a.Unpack {
		return ""
	}
	switch v := syntax.UnwrapParens(a.Value).(type) {
	case *syntax.Literal:
		if v.Raw == "0" {
			return "0"
		}
	case *syntax.ConstFetch:
		return strings.ToLower(strings.TrimPrefix(v.Name.Value, `\`))
	}
	return ""
}
