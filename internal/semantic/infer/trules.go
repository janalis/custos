package infer

import (
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

// TRules is the expression typer described by the "Expression type rules
// (T-rules)" of specs/UnnecessaryCasting.md (also referenced by
// CallableParameterUseCaseInTypeContext). It is layered over Env: rules it
// does not cover fall back to Env.TypeOf.
//
// Results are type sets with unknown parts dropped: types.Unknown (no atoms)
// means "empty". Atoms are raw (not normalised by any rule-specific scheme):
// class names keep their leading backslash and `T[]` forms are kept.
// SpecOnly retains self/static/parent member contracts as written for the
// consuming rule's translation; ordinary mode binds their class contexts.
//
// It is kept separate from Env on purpose: its rules come from the specs
// (partial sets, spec-defined arithmetic and overrides, optional SpecOnly
// mode) and differ from the engine's typing; folding it into Env would
// change the engine's types or the two rules' findings.
type TRules struct {
	Env *Env
	// SpecOnly evaluates the T-rules as literally written, without the
	// general engine's enrichments (CallableParameterUseCaseInTypeContext,
	// whose spec prefers "unknown" when in doubt):
	//   - a variable's type comes only from its parameter types and plain
	//     `=` assignments (`=&` included, typed by T-rules); foreach
	//     bindings and compound assignments neither contribute nor hide
	//     earlier assignments;
	//   - a method call's receiver is typed by T-rules only, and a receiver
	//     class lacking the method contributes nothing (the others still
	//     count);
	//   - the R-function subject argument is typed by T-rules;
	//   - arithmetic normalises `T[]` to array and sees a parenthesised
	//     numeric literal as a numeric literal;
	//   - `c ?: b` needs a non-empty type for `c` minus null.
	// Set it before the first TypeOf call.
	SpecOnly bool
	// DivisionIntOrFloat types `int / int` as int|float (PHP returns a float
	// unless the division is exact) instead of the spec heuristic's int
	// (UnnecessaryCasting). Set it before the first TypeOf call.
	DivisionIntOrFloat bool
	// SoundArithmetic types `+ - * / **` by PHP's actual result rules
	// instead of the spec heuristic (UnnecessaryCasting): an unknown
	// operand gives unknown, a float operand gives float, int with int
	// gives int (`/` and `**` with a non-literal exponent may give a
	// float), and any other operand (numeric strings, null, bool) gives
	// int|float. Set it before the first TypeOf call.
	SoundArithmetic bool
	cache           map[syntax.Expr]types.Type
	busy            map[syntax.Expr]bool
	assigns         map[syntax.Node]map[string][]*syntax.Assign // assignments, by scope
}

// NewTRules creates a T-rules typer over env.
func NewTRules(env *Env) *TRules {
	return &TRules{Env: env, cache: map[syntax.Expr]types.Type{}, busy: map[syntax.Expr]bool{}}
}

// KnownUnion unions the non-empty types (unknown parts are dropped).
func KnownUnion(ts ...types.Type) types.Type {
	var atoms []string
	for _, t := range ts {
		if !t.IsUnknown() {
			atoms = append(atoms, t.Atoms()...)
		}
	}
	if len(atoms) == 0 {
		return types.Unknown
	}
	return types.Of(atoms...)
}

// TypeOf returns the T-rules type of x.
func (r *TRules) TypeOf(x syntax.Expr) types.Type {
	if x == nil {
		return types.Unknown
	}
	if t, ok := r.cache[x]; ok {
		return t
	}
	if r.busy[x] {
		return types.Unknown
	}
	r.busy[x] = true
	t := r.infer(x)
	if !r.SpecOnly {
		switch x.(type) {
		case *syntax.MethodCall, *syntax.StaticCall:
			r.Env.TypeOf(x)
			if fact, ok := r.Env.chains[x]; ok {
				if fact.halted {
					t = types.Null
				} else {
					t = types.Union(t, types.Null)
				}
			}
		}
	}
	delete(r.busy, x)
	r.cache[x] = t
	return t
}

func (r *TRules) infer(x syntax.Expr) types.Type {
	e := r.Env
	switch n := x.(type) {
	case *syntax.FuncCall:
		if isFirstClassCallable(n.Args) {
			return e.TypeOf(n)
		}
	case *syntax.MethodCall:
		if isFirstClassCallable(n.Args) {
			return e.TypeOf(n)
		}
	case *syntax.StaticCall:
		if isFirstClassCallable(n.Args) {
			return e.TypeOf(n)
		}
	}
	switch n := x.(type) {
	case *syntax.Literal:
		return e.TypeOf(n)
	case *syntax.InterpolatedString:
		return types.String
	case *syntax.ConstFetch:
		return e.TypeOf(n)
	case *syntax.ClassConstFetch:
		return e.TypeOf(n)
	case *syntax.Paren:
		return r.TypeOf(n.Expr)
	case *syntax.Assign:
		if n.Op.Kind == syntax.TEqual {
			return r.TypeOf(n.Value)
		}
		return e.TypeOf(n)
	case *syntax.Unary:
		switch n.Op.Kind {
		case syntax.TMinus, syntax.TTilde:
			return r.TypeOf(n.Expr)
		}
		return e.TypeOf(n)
	case *syntax.Binary:
		switch n.Op.Kind {
		case syntax.TPlus, syntax.TMinus, syntax.TMul, syntax.TDiv, syntax.TPow:
			return r.arithmetic(n)
		case syntax.TCoalesce:
			l, rt := r.TypeOf(n.Left), r.TypeOf(n.Right)
			if l.IsUnknown() || rt.IsUnknown() {
				return types.Unknown
			}
			return KnownUnion(l.Without("null"), rt)
		}
		return e.TypeOf(n)
	case *syntax.Ternary:
		if n.Then == nil {
			// The left operand is the result only when truthy: never null
			// or false.
			l, rt := r.TypeOf(n.Cond), r.TypeOf(n.Else)
			if r.SpecOnly {
				l = l.Without("null", "false")
			}
			if !l.IsUnknown() && !rt.IsUnknown() {
				return KnownUnion(l.Without("null", "false"), rt)
			}
		} else {
			a, b := r.TypeOf(n.Then), r.TypeOf(n.Else)
			if !a.IsUnknown() && !b.IsUnknown() {
				return KnownUnion(a, b)
			}
		}
		return e.TypeOf(n)
	case *syntax.ArrayDimFetch:
		if t, ok := superglobalDim(n); ok {
			return t
		}
		return e.TypeOf(n)
	case *syntax.PropertyFetch:
		if v, ok := n.Var.(*syntax.Variable); ok && v.Name == "this" {
			if id, ok := n.Name.(*syntax.Identifier); ok {
				if cls := e.selfClass(syntax.EnclosingClass(n)); cls != "" {
					if p := e.Index.FindProperty(cls, id.Value, e.PHP); p != nil && p.Type != "" {
						return r.bindMember(types.FromDoc(p.Type, nil), p.Class, p.TypeClass, cls)
					}
				}
			}
		}
		return e.TypeOf(n)
	case *syntax.FuncCall:
		return r.funcCall(n)
	case *syntax.MethodCall:
		id, ok := n.Name.(*syntax.Identifier)
		if !ok {
			return types.Unknown
		}
		recv := r.TypeOf(n.Var)
		if recv.IsUnknown() && !r.SpecOnly {
			recv = e.TypeOf(n.Var)
		}
		var ts []types.Type
		for _, cls := range recv.Classes() {
			m := e.Index.FindMethod(strings.TrimPrefix(cls, `\`), id.Value, e.PHP)
			if m == nil {
				if r.SpecOnly {
					continue
				}
				return types.Unknown
			}
			ts = append(ts, r.bindMember(r.declaredAndDoc(m.Return, m.DocReturn, m.Builtin), m.Class, m.TypeClass, strings.TrimPrefix(cls, `\`)))
		}
		return KnownUnion(ts...)
	case *syntax.StaticCall:
		id, ok := n.Name.(*syntax.Identifier)
		if !ok {
			return types.Unknown
		}
		cls := e.classRef(n.Class)
		if cls == "" {
			return types.Unknown
		}
		m := e.Index.FindMethod(cls, id.Value, e.PHP)
		if m == nil {
			return types.Unknown
		}
		receiver := cls
		if !r.SpecOnly {
			if nm, ok := n.Class.(*syntax.Name); ok && strings.EqualFold(nm.Value, "parent") {
				if caller := e.selfClass(syntax.EnclosingClass(n)); caller != "" {
					receiver = caller
				}
			}
		}
		return r.bindMember(r.declaredAndDoc(m.Return, m.DocReturn, m.Builtin), m.Class, m.TypeClass, receiver)
	case *syntax.Variable:
		return r.variable(n)
	}
	return e.TypeOf(x)
}

func (r *TRules) bindMember(t types.Type, declaration, effective, receiver string) types.Type {
	if r.SpecOnly {
		return t
	}
	return r.Env.bindMember(t, declaration, effective, receiver)
}

// DeclaredAndDoc unions a declared type string and a doc type string (both
// as stored in the index).
func DeclaredAndDoc(declared, doc string) types.Type {
	return KnownUnion(types.FromDoc(declared, nil), types.FromDoc(doc, nil))
}

func (r *TRules) floatish(s types.Type) bool {
	return s.IsUnknown() || s.Has("float") || s.Has("number") || (s.Has("string") && !s.Has("int"))
}

// hasArray reports whether s contains array (in SpecOnly mode, also any
// `T[]` form, as the arithmetic rule normalises its operands).
func (r *TRules) hasArray(s types.Type) bool {
	if s.Has("array") {
		return true
	}
	if r.SpecOnly {
		for _, a := range s.Atoms() {
			if strings.Contains(a, "[]") {
				return true
			}
		}
	}
	return false
}

func (r *TRules) arithmetic(n *syntax.Binary) types.Type {
	if r.SoundArithmetic {
		return r.soundArithmetic(n)
	}
	l := r.TypeOf(n.Left)
	isFloat := r.floatish(l)
	isArray := r.hasArray(l)
	if !isFloat || (!isArray && n.Op.Kind == syntax.TPlus) {
		rt := r.TypeOf(n.Right)
		isFloat = isFloat || r.floatish(rt)
		right := n.Right
		if r.SpecOnly {
			right = syntax.UnwrapParens(right)
		}
		isArray = (isArray && !isNumericLiteral(right)) || r.hasArray(rt)
	}
	switch {
	case isArray:
		return types.Array
	case isFloat:
		return types.Float
	case r.DivisionIntOrFloat && n.Op.Kind == syntax.TDiv:
		return types.Of("int", "float")
	}
	return types.Int
}

func isNumericLiteral(x syntax.Expr) bool {
	l, ok := x.(*syntax.Literal)
	return ok && (l.LitKind == syntax.LitInt || l.LitKind == syntax.LitFloat)
}

var superglobals = map[string]bool{
	"_GET": true, "_POST": true, "_COOKIE": true, "_REQUEST": true, "_SERVER": true,
	"_ENV": true, "_FILES": true, "_SESSION": true, "GLOBALS": true,
}

func superglobalDim(n *syntax.ArrayDimFetch) (types.Type, bool) {
	v, ok := n.Var.(*syntax.Variable)
	if !ok || !superglobals[v.Name] {
		return types.Unknown, false
	}
	if v.Name == "_SERVER" {
		if lit, ok := n.Dim.(*syntax.Literal); ok && lit.LitKind == syntax.LitString && len(lit.Raw) >= 2 {
			switch lit.Raw[1 : len(lit.Raw)-1] {
			case "argv":
				return types.Array, true
			case "argc", "REQUEST_TIME":
				// (REMOTE_PORT and SERVER_PORT are strings under web SAPIs.)
				return types.Int, true
			case "REQUEST_TIME_FLOAT":
				return types.Float, true
			}
			return types.String, true
		}
	}
	return types.Of("string", "array"), true
}

// ---- functions ------------------------------------------------------------------------

func (r *TRules) funcCall(n *syntax.FuncCall) types.Type {
	f := r.Env.ResolveFunction(n)
	if f == nil {
		return types.Unknown
	}
	if t, ok := r.override(n, f); ok {
		return t
	}
	t := r.declaredAndDoc(f.Return, f.DocReturn, f.Builtin)
	if st, ok := r.Env.safeCallType(f, n, t); ok {
		return st // thecodingmachine/safe wrappers
	}
	return t
}

// declaredAndDoc is DeclaredAndDoc, without the doc type of a project
// declaration over a native Env (see Env.Native).
func (r *TRules) declaredAndDoc(declared, doc string, builtin bool) types.Type {
	if r.Env.userDoc(builtin) {
		return types.FromDoc(declared, nil)
	}
	if builtin && declared != "" {
		// The version-resolved signature is authoritative; the stub doc
		// only refines it (see builtinMemberType).
		d := types.FromDoc(declared, nil)
		return KnownUnion(d, refining(d, types.FromDoc(doc, nil)))
	}
	return DeclaredAndDoc(declared, doc)
}

// arg returns the argument of n (a call of f) for f's parameter i, positional
// or named; nil when absent. The parser always gives a FuncCall an argument
// list (parseArgs), so n.Args is never nil here.
func (r *TRules) arg(n *syntax.FuncCall, f *index.Function, i int) syntax.Expr {
	if i < len(n.Args.Args) {
		if a, ok := n.Args.Args[i].(*syntax.Arg); ok && a.Name == nil && !a.Unpack {
			return a.Value
		}
	}
	// Named argument: match the callee's parameter name at position i (the
	// stubs declare every parameter the overrides read; the bound only
	// guards against a stub lacking one).
	if i >= len(f.Params) {
		return nil
	}
	want := strings.TrimPrefix(f.Params[i].Name, "$")
	for _, x := range n.Args.Args {
		if a, ok := x.(*syntax.Arg); ok && a.Name != nil && strings.EqualFold(a.Name.Value, want) {
			return a.Value
		}
	}
	return nil
}

func (r *TRules) argCount(n *syntax.FuncCall) int {
	return len(n.Args.Args)
}

// override implements the R-function return types.
func (r *TRules) override(n *syntax.FuncCall, f *index.Function) (types.Type, bool) {
	switch name := strings.ToLower(strings.TrimPrefix(f.FQN, `\`)); name {
	case "str_replace", "str_ireplace", "preg_replace", "preg_replace_callback", "substr_replace",
		"preg_filter", "preg_replace_callback_array":
		idx := 2
		switch name {
		case "substr_replace":
			idx = 0
		case "preg_replace_callback_array":
			idx = 1
		}
		if s := r.arg(n, f, idx); s != nil {
			st := r.Env.TypeOf(s)
			if r.SpecOnly {
				// The spec types the subject by T-rules; what they leave
				// unknown the engine may still know (custos).
				if tt := r.TypeOf(s); !tt.IsUnknown() {
					st = tt
				}
			}
			if t, ok := replaceResult(st, strings.HasPrefix(name, "preg_")); ok {
				return t, true
			}
			if st.IsUnknown() && !r.SpecOnly {
				// Outside the spec-literal mode an unknown subject leaves
				// the result unknown (string or array).
				return types.Unknown, true
			}
		}
		res := []string{"string", "array"}
		if strings.HasPrefix(name, "preg_") { // null on a PCRE failure
			res = append(res, "null")
		}
		return types.Of(res...), true
	case "mb_convert_encoding": // custos: an array only for an array input
		if s := r.arg(n, f, 0); s != nil {
			st := r.Env.TypeOf(s)
			if r.SpecOnly {
				st = r.TypeOf(s)
			}
			switch {
			case st.IsUnknown():
			case st.OnlyOf("string", "int", "float", "bool", "true", "false", "null"):
				return types.Of("string", "bool"), true
			case st.IsArrayLike():
				return types.Of("array", "bool"), true
			}
		}
	case "strstr":
		return types.Of("string", "bool"), true
	case "get_class":
		return types.String, true
	case "explode":
		if r.argCount(n) >= 2 {
			if lit, ok := r.arg(n, f, 0).(*syntax.Literal); ok && lit.LitKind == syntax.LitString {
				if len(lit.Raw) <= 2 {
					return types.Bool, true
				}
				return types.Array, true
			}
		}
		return types.Of("array", "bool"), true
	case "parse_url":
		if r.argCount(n) == 2 {
			if c, ok := r.arg(n, f, 1).(*syntax.ConstFetch); ok {
				if strings.EqualFold(strings.TrimPrefix(c.Name.Value, `\`), "PHP_URL_PORT") {
					return types.Of("int", "null"), true
				}
				return types.Of("string", "null"), true
			}
		}
		return types.Of("array", "bool"), true
	case "current", "reset", "next", "prev", "end":
		return types.Mixed, true
	case "microtime":
		if r.argCount(n) == 1 {
			if c, ok := r.arg(n, f, 0).(*syntax.ConstFetch); !ok || !strings.EqualFold(strings.TrimPrefix(c.Name.Value, `\`), "false") {
				return types.Float, true
			}
		}
		return types.Int, true
	case "var_export", "print_r":
		if t, ok := printReturn(name, r.arg(n, f, 1), r.argCount(n)); ok {
			return t, true
		}
	case "pathinfo", "gettimeofday":
		// Argument-decided returns: the engine's override.
		if t := r.Env.TypeOf(n); !t.IsUnknown() {
			return t, true
		}
	case "max", "min":
		if r.argCount(n) >= 2 {
			var ts []types.Type
			for i := range r.argCount(n) {
				a, ok := n.Args.Args[i].(*syntax.Arg)
				if !ok || a.Unpack || a.Name != nil {
					return types.Unknown, false
				}
				ts = append(ts, r.TypeOf(a.Value))
			}
			if u := types.Union(ts...); !u.IsUnknown() && !u.Has("mixed") {
				return u.WithoutArrayInfo(), true
			}
		}
	case "abs":
		if r.argCount(n) == 1 {
			if a := r.arg(n, f, 0); a != nil {
				if t := r.TypeOf(a); t.OnlyOf("int", "float") {
					return t, true
				}
			}
		}
	}
	return types.Unknown, false
}

// ---- variables ------------------------------------------------------------------------

func (r *TRules) variable(v *syntax.Variable) types.Type {
	if v.Name == "" || v.NameExpr != nil {
		return types.Unknown
	}
	if v.Name == "this" {
		return r.Env.TypeOf(v)
	}
	scope := syntax.EnclosingVariableScope(v)
	if h, ok := scope.(*syntax.PropertyHook); ok && len(h.Params) == 0 && strings.EqualFold(h.Name.Value, "set") && v.Name == "value" {
		return r.Env.TypeOf(v)
	}
	pos := v.Span().Start
	// Outside SpecOnly, only the definitions reaching v count: an
	// unconditional reassignment (`$s = undeclared($s);`, of unknown type)
	// hides earlier ones instead of leaving their types in the union.
	reach, docs := r.reachingDefs(scope, v)
	ts := docs
	for _, p := range syntax.VariableScopeParams(scope) {
		if p.Var == nil || p.Var.Name != v.Name {
			continue
		}
		if reach != nil && !reach[p.Span().Start] {
			continue
		}
		if p.Variadic {
			return types.Array
		}
		ts = append(ts, r.ParamTypes(scope, p))
	}
	// A foreach enclosing v that rebinds the variable as its key/value hides
	// every earlier definition: only its binding and the definitions inside
	// its body are visible.
	var cutoff uint32
	unknownDef := false
	for p := v.Parent(); p != nil && p != scope && !r.SpecOnly; p = p.Parent() {
		fe, ok := p.(*syntax.Foreach)
		if !ok || fe.Body == nil || !containsPos(fe.Body, pos) {
			continue
		}
		if b := foreachBinding(fe, v.Name); b != nil {
			cutoff = fe.Span().Start
			bt := r.Env.TypeOf(b)
			ts = append(ts, bt)
			// An unknown element that reaches v keeps the union unknown,
			// like an unknown assignment below.
			unknownDef = bt.IsUnknown() && reach != nil && reach[b.Span().Start]
			break
		}
	}
	as := r.assignments(scope, v.Name)
	if len(as) > maxVarDefs {
		return types.Unknown // see maxVarDefs
	}
	after := uint32(0) // conditions before the last assignment do not narrow it
	for _, d := range as {
		later := d.Span().Start >= pos && !(reach != nil && reach[d.Span().Start])
		if later || d.Span().Start < cutoff || containsPos(d, pos) {
			continue
		}
		if reach != nil && !reach[d.Span().Start] {
			continue
		}
		if reach == nil && r.Env.exited(d.Span().Start, v, scope) {
			continue // SpecOnly: an assignment followed by return/throw/exit/break
		}
		after = max(after, d.Span().End)
		var dt types.Type
		switch {
		case d.Op.Kind == syntax.TEqual && (!d.ByRef || r.SpecOnly):
			dt = r.TypeOf(d.Value)
		case !r.SpecOnly:
			dt = r.Env.TypeOf(d)
		}
		ts = append(ts, dt)
		unknownDef = unknownDef || (reach != nil && dt.IsUnknown())
	}
	// A reaching definition of unknown type makes the variable unknown:
	// dropping it would leave a partial set that rules act on (`$f =
	// $o->x; if ($c) { $f = 'x'; } (string) $f`). Only with reaching
	// definitions (never in SpecOnly mode, whose sets are partial by spec).
	if unknownDef || r.unmodelled(scope, v, reach, as) {
		return types.Unknown
	}
	clob, undef := r.Env.dynamicRead(v)
	if clob {
		return types.Unknown // extract(), parse_str(), `$$name =` since
	}
	if undef && !r.SpecOnly {
		ts = append(ts, types.Null)
	}
	t := KnownUnion(ts...)
	if !t.IsUnknown() {
		// Guards on the path (`if (null !== $v)`, is_*(), instanceof…).
		// T-rules sets may be partial (unknown members dropped), so a guard
		// is only trusted to remove members: a result naming types outside
		// the known set comes from the unknown part.
		n := r.Env.narrow(t, v, scope, after)
		for _, a := range n.Atoms() {
			if !t.Has(a) {
				return types.Unknown
			}
		}
		t = n
	}
	return t
}

// unmodelled reports whether a definition reaching v is of a kind the
// T-rules do not type: destructuring (`list($h, $m) = explode(…)`), out
// arguments, catch, global/static, a foreach binding read after its loop,
// a by-reference closure import. Typing v from the other definitions
// alone would give a partial set the casting rule trusts (outside
// SpecOnly; reach nil: SpecOnly).
func (r *TRules) unmodelled(scope syntax.Node, v *syntax.Variable, reach map[uint32]bool, as []*syntax.Assign) bool {
	if reach == nil {
		return false
	}
	modelled := make(map[uint32]bool, len(as)+4)
	for _, p := range syntax.VariableScopeParams(scope) {
		modelled[p.Span().Start] = true
	}
	for _, d := range as {
		modelled[d.Span().Start] = true
	}
	for _, d := range r.Env.scopeVars(scope).defs[v.Name] {
		if d.doc {
			modelled[d.pos] = true
		}
	}
	for p := v.Parent(); p != nil && p != scope; p = p.Parent() {
		if fe, ok := p.(*syntax.Foreach); ok && fe.Body != nil && containsPos(fe.Body, v.Span().Start) {
			if b := foreachBinding(fe, v.Name); b != nil {
				modelled[b.Span().Start] = true
			}
		}
	}
	for p := range reach {
		if !modelled[p] {
			return true
		}
	}
	return false
}

// foreachBinding returns the variable named name bound by fe's key or value
// (including list destructuring), or nil.
// reachingDefs returns the positions of the definitions of v (Env's
// reaching definitions, forward ones) that reach it, and the types of the
// reaching inline `@var` annotations that annotate no assignment; reach is
// nil in SpecOnly mode or when they cannot be computed (no definition
// known to Env, beyond maxVarDefs), in which case every earlier
// assignment counts.
func (r *TRules) reachingDefs(scope syntax.Node, v *syntax.Variable) (reach map[uint32]bool, docs []types.Type) {
	if r.SpecOnly {
		return nil, nil
	}
	defs := r.Env.scopeVars(scope).defs[v.Name]
	if len(defs) == 0 || len(defs) > maxVarDefs {
		return nil, nil
	}
	fwd, back, _ := r.Env.reaching(defs, v, scope)
	if _, ok := r.Env.dynReads[v]; !ok && len(fwd)+len(back) > 0 {
		// The dynamic-write facts of this read, from the same reaching
		// definitions (dynamicRead then reads the cache).
		sv := r.Env.scopeVars(scope)
		clob := r.Env.clobbered(sv, fwd, back, v, scope)
		r.Env.noteDynRead(v, clob, !clob && r.Env.maybeUndefined(sv, defs, fwd, v, scope))
	}
	reach = make(map[uint32]bool, len(fwd)+len(back))
	for _, d := range fwd {
		reach[d.pos] = true
	}
	// Definitions later in a loop around v reach it on the next iteration
	// (`foreach (…) { (int) $n; $n = f(); }`).
	for _, d := range back {
		reach[d.pos] = true
	}
	for i, d := range defs {
		if !d.doc || !reach[d.pos] {
			continue
		}
		// `/** @var T $x */ $x = v;`: reaching keeps the annotation in
		// place of the assignment it annotates, whose value the T-rules
		// still use; a standalone annotation states the type itself.
		if i+1 < len(defs) && !defs[i+1].doc && defs[i+1].w == nil && !r.Env.semicolonBetween(d.docEnd, defs[i+1].pos) {
			reach[defs[i+1].pos] = true
		} else {
			docs = append(docs, d.typ())
		}
	}
	return reach, docs
}

func foreachBinding(fe *syntax.Foreach, name string) *syntax.Variable {
	var found *syntax.Variable
	for _, e := range []syntax.Expr{fe.Key, fe.Value} {
		if e == nil || found != nil {
			continue
		}
		syntax.Inspect(e, func(n syntax.Node) bool {
			if b, ok := n.(*syntax.Variable); ok && b.NameExpr == nil && b.Name == name && found == nil {
				found = b
			}
			return found == nil
		})
	}
	return found
}

func containsPos(n syntax.Node, pos uint32) bool {
	s := n.Span()
	return s.Start <= pos && pos < s.End
}

// assignments lists the assignments (plain and compound) whose target is the
// variable `name`, directly in scope (nil scope: the file's top level).
// The scope is walked once (cached): walking it for every read was
// quadratic on long bodies.
func (r *TRules) assignments(scope syntax.Node, name string) []*syntax.Assign {
	if byName, ok := r.assigns[scope]; ok {
		return byName[name]
	}
	byName := map[string][]*syntax.Assign{}
	visit := func(root syntax.Node) {
		syntax.Inspect(root, func(n syntax.Node) bool {
			switch n := n.(type) {
			case *syntax.Function, *syntax.Method, *syntax.Closure, *syntax.ArrowFunction, *syntax.PropertyHook, *syntax.ClassLike:
				return n == scope
			case *syntax.Assign:
				if t, ok := n.Var.(*syntax.Variable); ok && t.NameExpr == nil {
					byName[t.Name] = append(byName[t.Name], n)
				}
			}
			return true
		})
	}
	if scope == nil {
		for _, st := range r.Env.File.Stmts {
			visit(st)
		}
	} else {
		visit(scope)
	}
	if r.assigns == nil {
		r.assigns = map[syntax.Node]map[string][]*syntax.Assign{}
	}
	r.assigns[scope] = byName
	return byName[name]
}

// ParamTypes returns the declared type of a parameter united with its
// @param docblock types (no default-value contribution). Unlike
// Env.paramType it neither wraps variadics nor adds null for a `= null`
// default, as the T-rules specify.
func (r *TRules) ParamTypes(scope syntax.Node, p *syntax.Param) types.Type {
	at := p.Span().Start
	res := r.Env.resolver(at)
	declared := types.FromNode(p.Type, res)
	var doc types.Type
	if d := r.Env.DocOf(scope); d != nil && p.Var != nil && !r.Env.native {
		for _, dp := range d.EffectiveParams() {
			if dp.Name == p.Var.Name {
				doc = types.FromDoc(dp.Type, res)
			}
		}
	}
	return KnownUnion(declared, doc)
}

// soundArithmetic implements SoundArithmetic.
func (r *TRules) soundArithmetic(n *syntax.Binary) types.Type {
	l, rt := r.TypeOf(n.Left), r.TypeOf(n.Right)
	if l.IsUnknown() || rt.IsUnknown() {
		return types.Unknown
	}
	// Array operands (`array` or element-typed `T[]`): only array + array
	// is an array; anything else throws a TypeError.
	if hasArrayAtom(l) || hasArrayAtom(rt) {
		if n.Op.Kind == syntax.TPlus && l.IsArrayLike() && rt.IsArrayLike() {
			return types.Array
		}
		return types.Unknown
	}
	intOrFloat := types.Of("int", "float")
	switch {
	case l.OnlyOf("float") || rt.OnlyOf("float"):
		return types.Float
	case l.OnlyOf("int") && rt.OnlyOf("int"):
		switch n.Op.Kind {
		case syntax.TDiv:
			return intOrFloat
		case syntax.TPow:
			if lit, ok := syntax.UnwrapParens(n.Right).(*syntax.Literal); !ok || lit.LitKind != syntax.LitInt {
				return intOrFloat // a negative exponent gives a float
			}
		}
		return types.Int
	}
	return intOrFloat
}

// hasArrayAtom reports an `array` or `T[]` atom in s.
func hasArrayAtom(s types.Type) bool {
	for _, a := range s.Atoms() {
		if a == "array" || strings.HasSuffix(a, "[]") {
			return true
		}
	}
	return false
}
