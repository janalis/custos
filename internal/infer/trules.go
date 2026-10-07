package infer

import (
	"strings"

	"custos/internal/index"
	"custos/internal/phpdoc"
	"custos/internal/syntax"
	"custos/internal/types"
)

// TRules is the expression typer described by the "Expression type rules
// (T-rules)" of specs/UnnecessaryCasting.md (also referenced by
// CallableParameterUseCaseInTypeContext). It is layered over Env: rules it
// does not cover fall back to Env.TypeOf.
//
// Results are type sets with unknown parts dropped: types.Unknown (no atoms)
// means "empty". Atoms are raw (not normalised by any rule-specific scheme):
// class names keep their leading backslash, `self`/`static` returned by
// methods are kept as written, `T[]` forms are kept.
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
	cache              map[syntax.Expr]types.Type
	busy               map[syntax.Expr]bool
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
	delete(r.busy, x)
	r.cache[x] = t
	return t
}

func (r *TRules) infer(x syntax.Expr) types.Type {
	e := r.Env
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
				if cls := e.ClassFQN(EnclosingClass(n)); cls != "" {
					if p := e.Index.FindProperty(cls, id.Value, e.PHP); p != nil && p.Type != "" {
						return types.FromDoc(p.Type, nil)
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
			ts = append(ts, DeclaredAndDoc(m.Return, m.DocReturn))
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
		return DeclaredAndDoc(m.Return, m.DocReturn)
	case *syntax.Variable:
		return r.variable(n)
	}
	return e.TypeOf(x)
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
	l := r.TypeOf(n.Left)
	isFloat := r.floatish(l)
	isArray := r.hasArray(l)
	if !isFloat || (!isArray && n.Op.Kind == syntax.TPlus) {
		rt := r.TypeOf(n.Right)
		isFloat = isFloat || r.floatish(rt)
		right := n.Right
		if r.SpecOnly {
			right = unparen(right)
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
			case "argc", "REQUEST_TIME", "REMOTE_PORT", "SERVER_PORT":
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
	if t, ok := r.override(n, strings.ToLower(strings.TrimPrefix(f.FQN, `\`))); ok {
		return t
	}
	return DeclaredAndDoc(f.Return, f.DocReturn)
}

func (r *TRules) arg(n *syntax.FuncCall, i int) syntax.Expr {
	if n.Args == nil {
		return nil
	}
	if i < len(n.Args.Args) {
		if a, ok := n.Args.Args[i].(*syntax.Arg); ok && a.Name == nil && !a.Unpack {
			return a.Value
		}
	}
	// Named argument: match the callee's parameter name at position i.
	f := r.Env.ResolveFunction(n)
	if f == nil || i >= len(f.Params) {
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
	if n.Args == nil {
		return 0
	}
	return len(n.Args.Args)
}

// override implements the R-function return types.
func (r *TRules) override(n *syntax.FuncCall, name string) (types.Type, bool) {
	switch name {
	case "str_replace", "str_ireplace", "preg_replace", "preg_replace_callback", "substr_replace",
		"preg_filter", "preg_replace_callback_array":
		idx := 2
		switch name {
		case "substr_replace":
			idx = 0
		case "preg_replace_callback_array":
			idx = 1
		}
		res := []string{"string", "array"}
		if s := r.arg(n, idx); s != nil {
			st := r.Env.TypeOf(s)
			if r.SpecOnly {
				st = r.TypeOf(s)
			}
			if !st.IsUnknown() {
				hasArr := false
				for _, a := range st.Atoms() {
					if a == "array" || strings.HasSuffix(a, "[]") {
						hasArr = true
					}
				}
				res = res[:0]
				if st.Has("string") {
					res = append(res, "string")
				}
				if hasArr {
					res = append(res, "array")
				}
				if len(res) == 0 {
					return types.Unknown, true
				}
			}
		}
		if strings.HasPrefix(name, "preg_") { // null on a PCRE failure
			res = append(res, "null")
		}
		return types.Of(res...), true
	case "strstr":
		return types.Of("string", "bool"), true
	case "get_class":
		return types.String, true
	case "explode":
		if r.argCount(n) >= 2 {
			if lit, ok := r.arg(n, 0).(*syntax.Literal); ok && lit.LitKind == syntax.LitString {
				if len(lit.Raw) <= 2 {
					return types.Bool, true
				}
				return types.Array, true
			}
		}
		return types.Of("array", "bool"), true
	case "parse_url":
		if r.argCount(n) == 2 {
			if c, ok := r.arg(n, 1).(*syntax.ConstFetch); ok {
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
			if c, ok := r.arg(n, 0).(*syntax.ConstFetch); !ok || !strings.EqualFold(strings.TrimPrefix(c.Name.Value, `\`), "false") {
				return types.Float, true
			}
		}
		return types.Int, true
	case "abs":
		if r.argCount(n) == 1 {
			if a := r.arg(n, 0); a != nil {
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
	scope := scopeOf(v)
	pos := v.Span().Start
	var ts []types.Type
	for _, p := range scopeParams(scope) {
		if p.Var == nil || p.Var.Name != v.Name {
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
	for p := v.Parent(); p != nil && p != scope && !r.SpecOnly; p = p.Parent() {
		fe, ok := p.(*syntax.Foreach)
		if !ok || fe.Body == nil || !containsPos(fe.Body, pos) {
			continue
		}
		if b := foreachBinding(fe, v.Name); b != nil {
			cutoff = fe.Span().Start
			ts = append(ts, r.Env.TypeOf(b))
			break
		}
	}
	for _, d := range r.assignments(scope, v.Name) {
		if d.Span().Start >= pos || d.Span().Start < cutoff || containsPos(d, pos) {
			continue
		}
		switch {
		case d.Op.Kind == syntax.TEqual && (!d.ByRef || r.SpecOnly):
			ts = append(ts, r.TypeOf(d.Value))
		case !r.SpecOnly:
			ts = append(ts, r.Env.TypeOf(d))
		}
	}
	t := KnownUnion(ts...)
	if !t.IsUnknown() {
		// Guards on the path (`if (null !== $v)`, is_*(), instanceof…).
		// T-rules sets may be partial (unknown members dropped), so a guard
		// is only trusted to remove members: a result naming types outside
		// the known set comes from the unknown part.
		n := r.Env.narrow(t, v, scope)
		for _, a := range n.Atoms() {
			if !t.Has(a) {
				return types.Unknown
			}
		}
		t = n
	}
	return t
}

// foreachBinding returns the variable named name bound by fe's key or value
// (including list destructuring), or nil.
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

func scopeParams(scope syntax.Node) []*syntax.Param {
	switch s := scope.(type) {
	case *syntax.Function:
		return s.Params
	case *syntax.Method:
		return s.Params
	case *syntax.Closure:
		return s.Params
	case *syntax.ArrowFunction:
		return s.Params
	}
	return nil
}

// assignments lists the assignments (plain and compound) whose target is the
// variable `name`, directly in scope (nil scope: the file's top level).
func (r *TRules) assignments(scope syntax.Node, name string) []*syntax.Assign {
	var out []*syntax.Assign
	visit := func(root syntax.Node) {
		syntax.Inspect(root, func(n syntax.Node) bool {
			switch n := n.(type) {
			case *syntax.Function, *syntax.Method, *syntax.Closure, *syntax.ArrowFunction, *syntax.ClassLike:
				return n == scope
			case *syntax.Assign:
				if t, ok := n.Var.(*syntax.Variable); ok && t.NameExpr == nil && t.Name == name {
					out = append(out, n)
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
	return out
}

// ParamTypes returns the declared type of a parameter united with its
// @param docblock types (no default-value contribution).
func (r *TRules) ParamTypes(scope syntax.Node, p *syntax.Param) types.Type {
	at := p.Span().Start
	res := r.Env.resolver(at)
	declared := types.FromNode(p.Type, res)
	var doc types.Type
	if c := index.DocComment(r.Env.File, scope); c != "" && p.Var != nil {
		for _, dp := range phpdoc.Parse(c).Params() {
			if dp.Name == p.Var.Name {
				doc = types.FromDoc(dp.Type, res)
			}
		}
	}
	return KnownUnion(declared, doc)
}
