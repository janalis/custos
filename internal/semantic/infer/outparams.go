package infer

import (
	"cmp"
	"slices"
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/stubs"
	"custos/internal/semantic/types"
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
	outInt        outKind = iota + 1
	outFloat              // float
	outString             // string
	outArray              // array of unknown elements
	outMatches            // preg_match: string[] without flags, else array
	outMatchesAll         // preg_match_all: string[][] without flags, else array
	outReindex
	outPreserve
	outPush
	outUnshift
	outSplice
	outSettype
	outAppendLines // exec: string lines appended to an array
	outUnchanged   // the variable keeps its type: an array whose pointer moves or that loses an end
	outElements    // an array keeps its keys, its elements may change
)

// byRefOut is the documented type a user callee leaves in by-reference
// parameter p: its `@param-out`, else its `@param` (the callee is trusted to
// write back what it documents taking); "" when neither is documented.
func byRefOut(p index.Param) string {
	return cmp.Or(p.Out, p.DocType)
}

// noBuiltinOut distinguishes a known builtin without an output contract from a
// user callee. It is read-only and shared across calls.
var noBuiltinOut = map[int]outKind{}

// builtinOut lists, per builtin function, the output parameters by
// position and the value they hold after the call.
var builtinOut = map[string]map[int]outKind{
	"sort": {0: outReindex}, "rsort": {0: outReindex}, "usort": {0: outReindex}, "shuffle": {0: outReindex},
	"ksort": {0: outPreserve}, "krsort": {0: outPreserve}, "asort": {0: outPreserve}, "arsort": {0: outPreserve}, "uasort": {0: outPreserve}, "uksort": {0: outPreserve}, "natsort": {0: outPreserve}, "natcasesort": {0: outPreserve},
	"array_push": {0: outPush}, "array_unshift": {0: outUnshift}, "array_splice": {0: outSplice}, "settype": {0: outSettype},
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
	"end":                         {0: outUnchanged},
	"reset":                       {0: outUnchanged},
	"next":                        {0: outUnchanged},
	"prev":                        {0: outUnchanged},
	"array_pop":                   {0: outUnchanged},
	"array_shift":                 {0: outUnchanged},
	"array_walk":                  {0: outElements},
	"array_walk_recursive":        {0: outElements},
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
		if method, ok := call.(*syntax.MethodCall); ok {
			e.collectDeferredMethodOut(method, sv)
		}
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
				typ = e.builtinOutType(kind, al, v, params)
			} else {
				// Without a listed contract, the variable holds what the
				// parameter is declared to take; the stubs document PHP's
				// own types.
				declared := cmp.Or(p.Type, p.DocType)
				typ = func() types.Type {
					if t := e.docType(declared, nil); declared != "" && !t.Has("mixed") {
						return t
					}
					return types.Unknown
				}
			}
		} else if out := byRefOut(p); out != "" && !e.native {
			typ = func() types.Type {
				if t := e.docType(out, nil); !t.Has("mixed") {
					return t
				}
				return types.Unknown
			}
		}
		if typ == nil && builtin == nil {
			typ = func() types.Type { return types.Unknown }
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
		if e.optionalAssertion(call) {
			kill = syntax.Span{}
			typ = func() types.Type { return types.Mixed }
		}
		arg := &mutation{span: call.Span(), kind: mutCallArg, call: call, arg: a, idx: i}
		sv.defs[v.Name] = append(sv.defs[v.Name], varDef{pos: call.Span().Start, end: call.Span().End, kill: kill, typ: typ, arg: arg})
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
			return f.Params, noBuiltinOut
		}
		return f.Params, nil
	case *syntax.MethodCall:
		if e.hasNullsafeChain(c) {
			return nil, nil
		}
		id, ok := c.Name.(*syntax.Identifier)
		recv, isVar := syntax.UnwrapParens(c.Var).(*syntax.Variable)
		if !ok || !isVar || recv.NameExpr != nil {
			return nil, nil
		}
		cls := ""
		if recv.Name == "this" {
			cls = e.selfClass(syntax.EnclosingClass(c))
		} else {
			scope := syntax.EnclosingVariableScope(c)
			body := syntax.VariableScopeBody(scope)
			if body == nil {
				return nil, nil
			}
			if sv := e.scopes[scope]; sv != nil {
				for _, d := range sv.defs[recv.Name] {
					if d.pos >= body.Span().Start && d.pos < c.Span().Start {
						return nil, nil
					}
				}
			}
			for _, p := range syntax.VariableScopeParams(scope) {
				if p.Var.Name == recv.Name {
					classes := types.FromNode(p.Type, e.resolver(p.Span().Start)).Classes()
					if len(classes) == 1 {
						cls = strings.TrimPrefix(classes[0], `\`)
					}
				}
			}
		}
		if cls != "" {
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
func (e *Env) builtinOutType(kind outKind, al *syntax.ArgList, v *syntax.Variable, params []index.Param) func() types.Type {
	switch kind {
	case outUnchanged:
		return nil
	case outElements:
		return func() types.Type {
			if prior := e.TypeOf(v); prior.IsArrayLike() {
				return types.Array.WithArrayKey(prior.ArrayKey())
			}
			return types.Unknown
		}
	case outSettype:
		return func() types.Type {
			value := e.tplArg(al, params, 1)
			if value == nil {
				return types.Unknown
			}
			lit, ok := syntax.UnwrapParens(value).(*syntax.Literal)
			if !ok || lit.LitKind != syntax.LitString {
				return types.Unknown
			}
			switch strings.ToLower(strings.Trim(lit.Raw, "'\"")) {
			case "boolean", "bool":
				return types.Bool
			case "integer", "int":
				return types.Int
			case "float", "double":
				return types.Float
			case "string":
				return types.String
			case "array":
				return types.Array
			case "object":
				return types.Of("object")
			case "null":
				return types.Null
			}
			return types.Unknown
		}
	case outReindex, outPreserve, outPush, outUnshift, outSplice:
		return func() types.Type {
			prior := e.TypeOf(v)
			if e.PHP != 0 && e.PHP < phpversion.PHP80 && !prior.IsArrayLike() {
				for _, atom := range prior.Atoms() {
					if atom == "array" || strings.HasSuffix(atom, "[]") {
						return types.Unknown
					}
				}
				return prior
			}
			elem := iterElem(prior)
			if kind == outPush || kind == outUnshift {
				for i, item := range al.Args {
					if a, ok := item.(*syntax.Arg); ok {
						if p, valid := paramFor(params, a, i); !valid || paramIndex(params, p) == 0 {
							continue
						}
						t := e.TypeOf(a.Value)
						if a.Unpack {
							t = iterElem(t)
						}
						elem = types.Union(elem, t)
					}
				}
			}
			if kind == outSplice {
				if replacement := e.tplArg(al, params, 3); replacement != nil {
					t := e.TypeOf(replacement)
					if t.IsArrayLike() {
						t = iterElem(t)
					}
					if !t.OnlyOf("null") {
						elem = types.Union(elem, t.Without("null"))
					}
				}
			}
			result := arrayOf(elem)
			if result.IsUnknown() {
				result = types.Array
			}
			if kind == outReindex {
				return result.WithArrayKey(types.Int)
			}
			key := prior.ArrayKey()
			if kind == outPush || kind == outUnshift || kind == outSplice {
				key = types.Union(key, types.Int)
			}
			return result.WithArrayKey(key)
		}
	case outInt:
		return func() types.Type { return types.Int }
	case outFloat:
		return func() types.Type { return types.Float }
	case outString:
		return func() types.Type { return types.String }
	case outMatches, outMatchesAll:
		// Unmatched groups are '' (trailing ones absent) without flags;
		// PREG_UNMATCHED_AS_NULL makes them null, and PREG_OFFSET_CAPTURE
		// pairs each match with its offset. preg_match_all() lists the
		// matches per group, or with PREG_SET_ORDER per match: strings
		// either way. Other flags: plain array.
		flags := "0"
		if len(al.Args) > 3 {
			flags = matchFlags(al.Args[3])
		}
		el := matchesElement(flags, kind == outMatchesAll)
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

// matchesElement returns the doc type of each group's entry in the matches
// of a preg function with the named flags (see matchFlags), "" when they
// are unknown: unmatched groups are ” without flags, PREG_UNMATCHED_AS_NULL
// makes them null, PREG_OFFSET_CAPTURE pairs each match with its offset;
// PREG_PATTERN_ORDER and PREG_SET_ORDER only order preg_match_all()'s.
func matchesElement(flags string, all bool) string {
	switch flags {
	case "0":
		return "string"
	case "preg_pattern_order", "preg_set_order":
		if all {
			return "string"
		}
	case "preg_unmatched_as_null":
		return "(string|null)"
	case "preg_offset_capture":
		return "array{0: string, 1: int}"
	}
	return ""
}

// matchFlags names the flags argument of preg_match(): "0" for a literal 0,
// the lower-case constant name for a single constant, "" otherwise.
func matchFlags(x syntax.Expr) string {
	a, ok := x.(*syntax.Arg)
	if !ok || a.Name != nil || a.Unpack {
		return ""
	}
	return flagsValue(a.Value)
}

// flagsValue is matchFlags for the value of the argument.
func flagsValue(x syntax.Expr) string {
	switch v := syntax.UnwrapParens(x).(type) {
	case *syntax.Literal:
		if v.Raw == "0" {
			return "0"
		}
	case *syntax.ConstFetch:
		return strings.ToLower(strings.TrimPrefix(v.Name.Value, `\`))
	}
	return ""
}

// collectDeferredMethodOut postpones receiver inference until all definitions
// have been collected. Arguments passed by value or to unresolved methods keep
// their type at the argument, before the call's possible write.
func (e *Env) collectDeferredMethodOut(call *syntax.MethodCall, sv *scopeVars) {
	id, ok := call.Name.(*syntax.Identifier)
	if !ok {
		return
	}
	kill := unconditionalBlock(call)
	if arrow := e.arrowDefinitionSpan(call); arrow.Len() > 0 {
		kill = arrow
	}
	for i, item := range call.Args.Args {
		a, ok := item.(*syntax.Arg)
		if !ok || a.Unpack {
			continue
		}
		v, ok := a.Value.(*syntax.Variable)
		if !ok || v.Name == "" || v.Name == "this" {
			continue
		}
		typ := func() types.Type {
			receiver := e.TypeOf(call.Var)
			classes := receiver.Classes()
			keepPrior := e.hasNullsafeChain(call) || receiver.HasAny("mixed", "object")
			// unresolved: a callee the call may reach is unknown.
			unresolved := len(classes) == 0 || receiver.HasAny("mixed", "object")
			var results []types.Type
			for _, cls := range classes {
				m := e.Index.FindMethod(strings.TrimPrefix(cls, `\`), id.Value, e.PHP)
				if m == nil {
					keepPrior, unresolved = true, true
					continue
				}
				p, valid := paramFor(m.Params, a, i)
				if !valid || !p.ByRef {
					keepPrior = true
					continue
				}
				if byRefOut(p) == "" || e.native {
					return types.Unknown
				}
				t := e.docType(byRefOut(p), nil)
				if t.Has("mixed") {
					return types.Unknown
				}
				results = append(results, t)
			}
			prior := e.TypeOf(v)
			if e.native && unresolved {
				// The callee may be unknown and take it by reference: what
				// conditions proved about it before the call does not hold.
				prior = e.definedType(v)
			}
			if len(results) == 0 {
				return prior
			}
			if keepPrior {
				results = append(results, prior)
			}
			return types.Union(results...)
		}
		if e.optionalAssertion(call) {
			kill = syntax.Span{}
			typ = func() types.Type { return types.Mixed }
		}
		arg := &mutation{span: call.Span(), kind: mutCallArg, call: call, arg: a, idx: i}
		sv.defs[v.Name] = append(sv.defs[v.Name], varDef{pos: call.Span().Start, end: call.Span().End, kill: kill, typ: typ, arg: arg})
	}
}

// definedType types read v by the definitions reaching it, without the
// conditions guarding it.
func (e *Env) definedType(v *syntax.Variable) types.Type {
	if e.definedBusy[v] {
		return types.Unknown // a cycle through a loop: nothing more is known
	}
	if e.definedBusy == nil {
		e.definedBusy = map[*syntax.Variable]bool{}
	}
	e.definedBusy[v] = true
	defer delete(e.definedBusy, v)
	scope := syntax.EnclosingVariableScope(v)
	sv := e.scopeVars(scope)
	fwd, back, _ := e.reaching(sv.defs[v.Name], v, scope)
	if len(fwd)+len(back) == 0 || e.clobbered(sv, fwd, back, v, scope) {
		return types.Unknown
	}
	ts := make([]types.Type, 0, len(fwd)+len(back))
	for _, d := range slices.Concat(fwd, back) {
		ts = append(ts, d.typ())
	}
	return types.Union(ts...)
}
