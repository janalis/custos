package infer

import (
	"strings"

	"custos/internal/syntax"
	"custos/internal/types"
)

// Comparison and type-guard narrowing: `===`/`==` and their negations
// (null/true/false literals, scalar literals and constants, `true ===
// cond`, `gettype($x) === 'string'`, `$x::class === Foo::class`), chains
// whose base must be non-null for the condition to hold (`$x?->m()`,
// `isset($x->p)`), `is_a()`, `is_subclass_of()` and strict `in_array()`.

// truthyType is t where it is truthy: neither null nor false, `true` for
// bool, a non-empty array.
func truthyType(t types.Type) types.Type {
	return replaceAtom(t.Without("null", "false"), "bool", "true").WithNonEmpty(true)
}

// replaceAtom replaces atom from of t by to (`bool` minus false is true).
func replaceAtom(t types.Type, from, to string) types.Type {
	if !t.Has(from) {
		return t
	}
	rest := t.Without(from)
	if len(rest.Atoms()) == 0 {
		return types.Of(to)
	}
	return types.Union(rest, types.Of(to))
}

// falsyType is t where it is falsy: `true` goes, bool becomes false.
// Other members may hold falsy values (empty string, 0, [], null) and stay.
func falsyType(t types.Type) types.Type {
	if !t.Has("true") && !t.Has("bool") {
		return t
	}
	return replaceAtom(t.Without("true"), "bool", "false")
}

// chainBaseNonNull narrows t, the type of name, where the property, method
// or element chain x on it (`$x->a`, `$x?->m()`, `$x['k']->b`) is known
// to be non-null: had the base been null (or false, for an element read)
// the chain would have been null (or thrown). x itself is not name.
func chainBaseNonNull(t types.Type, x syntax.Expr, name string) types.Type {
	r, _ := chainBase(t, x, name)
	return r
}

func chainBase(t types.Type, x syntax.Expr, name string) (types.Type, bool) {
	x = syntax.UnwrapParens(x)
	for {
		var base syntax.Expr
		dim := false
		switch n := x.(type) {
		case *syntax.PropertyFetch:
			base = n.Var
		case *syntax.MethodCall:
			base = n.Var
		case *syntax.ArrayDimFetch:
			base, dim = n.Var, true
		default:
			return t, false
		}
		base = syntax.UnwrapParens(base)
		if narrowKey(base) == name {
			out := t.Without("null")
			if dim {
				out = out.Without("false").WithNonEmpty(true)
			}
			if len(out.Atoms()) == 0 {
				return t, false
			}
			return out, true
		}
		x = base
	}
}

// instanceType is t narrowed to an instance of class cls (the generic
// arguments the type already had for it are kept).
func instanceType(t types.Type, cls string) types.Type {
	atom := `\` + cls
	if args := t.TypeArgs(atom); args != nil {
		return types.Of(atom).WithTypeArgs(atom, args)
	}
	return types.Of(atom)
}

// notInstance removes from t the class cls (when withSelf) and its
// subtypes (`string|PropertyPath` minus `instanceof PropertyPathInterface`).
func (e *Env) notInstance(t types.Type, cls string, withSelf bool) types.Type {
	var drop []string
	if withSelf {
		drop = append(drop, `\`+cls)
	}
	for _, a := range t.Classes() {
		if !strings.HasSuffix(a, "[]") && !strings.EqualFold(a, `\`+cls) &&
			e.Index.IsSubtype(strings.TrimPrefix(a, `\`), cls, e.PHP) {
			drop = append(drop, a)
		}
	}
	return t.Without(drop...)
}

// eqCond narrows t by comparison c (`===`, `!==`, `==`, `!=`).
func (e *Env) eqCond(t types.Type, c *syntax.Binary, name string, truthy bool, steps *int) types.Type {
	strict := c.Op.Kind == syntax.TIsIdentical || c.Op.Kind == syntax.TIsNotIdentical
	// holds: the operands are equal (identical when strict).
	holds := (c.Op.Kind == syntax.TIsIdentical || c.Op.Kind == syntax.TIsEqual) == truthy
	if isEmptyArrayComparison(c, name) {
		switch {
		case holds:
			return t
		case strict:
			return t.WithNonEmpty(true)
		}
		// `$x != []`: neither null, false nor an empty array.
		return t.Without("null", "false").WithNonEmpty(true)
	}
	if n, ok := countComparison(c, name); ok {
		return nonEmptyIf(t, countNonEmpty(c.Op.Kind, n, truthy))
	}
	l, r := syntax.UnwrapParens(c.Left), syntax.UnwrapParens(c.Right)
	switch {
	case isVar(l, name):
		return e.varCmp(t, r, strict, holds)
	case isVar(r, name):
		return e.varCmp(t, l, strict, holds)
	}
	// `true === is_string($x)`, `false == $x instanceof Foo`.
	for _, p := range [2][2]syntax.Expr{{l, r}, {r, l}} {
		lit := constLiteral(p[0])
		if (lit != "true" && lit != "false") || constLiteral(p[1]) != "" {
			continue
		}
		b := lit == "true"
		switch {
		case holds:
			// Equal to true is truthy, to false falsy (strictly or not).
			return e.applyCondB(t, p[1], name, b, steps)
		case !strict || isBoolExpr(p[1]):
			return e.applyCondB(t, p[1], name, !b, steps)
		}
		return t
	}
	for _, p := range [2][2]syntax.Expr{{l, r}, {r, l}} {
		if out, ok := e.typeFuncCmp(t, p[0], p[1], name, holds); ok {
			return out
		}
	}
	// `$x?->m() !== null`, `$x->kind === Kind::A`: the chain is not null.
	for _, p := range [2][2]syntax.Expr{{l, r}, {r, l}} {
		nullCmp := syntax.IsNullConst(p[1])
		if (nullCmp && !holds) || (!nullCmp && strict && holds && isConstValue(p[1])) {
			if out, ok := chainBase(t, p[0], name); ok {
				return out
			}
		}
	}
	return t
}

// varCmp narrows the type t of a variable compared with other.
func (e *Env) varCmp(t types.Type, other syntax.Expr, strict, holds bool) types.Type {
	lit := constLiteral(other)
	if !strict {
		if lit == "null" && !holds {
			return t.Without("null")
		}
		return t
	}
	if lit == "" {
		// `$x === 'a'`, `$x === Kind::A`: the constant's type.
		if holds {
			if vt, ok := e.constValueType(other); ok {
				return e.assertIs(t, vt)
			}
		}
		return t
	}
	if holds {
		if t.Has(lit) || (lit != "null" && t.Has("bool")) {
			return types.Of(lit)
		}
		return t
	}
	out := t.Without(lit)
	if lit != "null" {
		out = replaceAtom(out, "bool", map[string]string{"true": "false", "false": "true"}[lit])
	}
	return out
}

// isConstValue reports a literal or constant expression (whose value is
// not null unless it is the null constant).
func isConstValue(x syntax.Expr) bool {
	switch n := syntax.UnwrapParens(x).(type) {
	case *syntax.Literal, *syntax.ClassConstFetch:
		return true
	case *syntax.ConstFetch:
		return !syntax.IsNullConst(n)
	case *syntax.Unary:
		_, ok := syntax.UnwrapParens(n.Expr).(*syntax.Literal)
		return ok && (n.Op.Kind == syntax.TMinus || n.Op.Kind == syntax.TPlus)
	}
	return false
}

// constValueType is the type of a literal or constant expression x (false
// for anything else, or when the type is unknown or mixed).
func (e *Env) constValueType(x syntax.Expr) (types.Type, bool) {
	x = syntax.UnwrapParens(x)
	switch n := x.(type) {
	case *syntax.Literal:
		switch n.LitKind {
		case syntax.LitInt:
			return types.Of("int"), true
		case syntax.LitFloat:
			return types.Of("float"), true
		}
		return types.Of("string"), true
	case *syntax.Unary:
		if l, ok := syntax.UnwrapParens(n.Expr).(*syntax.Literal); ok && l.LitKind != syntax.LitString && (n.Op.Kind == syntax.TMinus || n.Op.Kind == syntax.TPlus) {
			return e.constValueType(l)
		}
		return types.Unknown, false
	case *syntax.ClassConstFetch, *syntax.ConstFetch:
		vt := e.TypeOf(x)
		if vt.IsUnknown() || vt.Has("mixed") {
			return types.Unknown, false
		}
		return vt.WithoutArrayInfo(), true
	}
	return types.Unknown, false
}

// isBoolExpr reports whether x always evaluates to a bool (so `x !== true`
// means x is false).
func isBoolExpr(x syntax.Expr) bool {
	switch n := syntax.UnwrapParens(x).(type) {
	case *syntax.Unary:
		return n.Op.Kind == syntax.TExclaim
	case *syntax.Binary:
		switch n.Op.Kind {
		case syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr, syntax.TXor,
			syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TIsEqual, syntax.TIsNotEqual,
			syntax.TLess, syntax.TIsSmallerOrEqual, syntax.TGreater, syntax.TIsGreaterOrEqual:
			return true
		}
	case *syntax.Instanceof, *syntax.Isset, *syntax.Empty:
		return true
	case *syntax.FuncCall:
		nm, ok := n.Name.(*syntax.Name)
		if !ok {
			return false
		}
		fn := strings.ToLower(strings.TrimPrefix(nm.Value, `\`))
		if _, ok := typeChecks[fn]; ok {
			return true
		}
		switch fn {
		case "in_array", "array_key_exists", "key_exists", "is_a", "is_subclass_of":
			return true
		}
	}
	return false
}

// gettypeNames maps gettype() results to the atoms they cover. A closed
// resource reads "resource (closed)", so "resource" only narrows when equal.
var gettypeNames = map[string][]string{
	"boolean": typeChecks["is_bool"], "integer": {"int"}, "double": {"float"}, "string": {"string"},
	"array": {"array"}, "object": {"object"}, "NULL": {"null"}, "resource": {"resource"},
}

// debugTypeNames maps the builtin get_debug_type() results to atoms.
var debugTypeNames = map[string][]string{
	"null": {"null"}, "bool": typeChecks["is_bool"], "int": {"int"}, "float": {"float"},
	"string": {"string"}, "array": {"array"},
}

// typeFuncCmp narrows t by `gettype($x)`, `get_debug_type($x)`,
// `get_class($x)` or `$x::class` (fx) compared with val; holds tells
// whether they are equal.
func (e *Env) typeFuncCmp(t types.Type, fx, val syntax.Expr, name string, holds bool) (types.Type, bool) {
	fn := ""
	switch n := fx.(type) {
	case *syntax.FuncCall:
		nm, ok := n.Name.(*syntax.Name)
		if !ok || len(n.Args.Args) != 1 {
			return t, false
		}
		a, ok := n.Args.Args[0].(*syntax.Arg)
		if !ok || a.Unpack || a.Name != nil || !isVar(a.Value, name) {
			return t, false
		}
		fn = strings.TrimPrefix(nm.Value, `\`)
		switch {
		case strings.EqualFold(fn, "gettype"), strings.EqualFold(fn, "get_debug_type"), strings.EqualFold(fn, "get_class"):
			fn = strings.ToLower(fn)
		default:
			return t, false
		}
	case *syntax.ClassConstFetch:
		id, ok := n.Name.(*syntax.Identifier)
		if !ok || !strings.EqualFold(id.Value, "class") || !isVar(n.Class, name) {
			return t, false
		}
		fn = "get_class"
	default:
		return t, false
	}
	s, isStr := "", false
	if l, ok := syntax.UnwrapParens(val).(*syntax.Literal); ok && l.LitKind == syntax.LitString {
		s, isStr = plainString(l.Raw)
	}
	switch fn {
	case "gettype":
		atoms, ok := gettypeNames[s]
		if !isStr || !ok {
			return t, true
		}
		if !holds && s == "resource" {
			return t, true
		}
		return narrowAtoms(t, atoms, holds), true
	case "get_debug_type":
		if atoms, ok := debugTypeNames[s]; isStr && ok {
			return narrowAtoms(t, atoms, holds), true
		}
	}
	// A class name: exactly that class (a subclass compares unequal, so
	// inequality removes nothing).
	cls := e.classNameValue(val)
	if cls == "" {
		return t, false
	}
	if holds {
		return instanceType(t, cls), true
	}
	return t, true
}

// classNameValue is the class named by `Foo::class` or a string literal
// holding a class name ("" otherwise).
func (e *Env) classNameValue(x syntax.Expr) string {
	switch n := syntax.UnwrapParens(x).(type) {
	case *syntax.ClassConstFetch:
		id, ok := n.Name.(*syntax.Identifier)
		if !ok || !strings.EqualFold(id.Value, "class") {
			return ""
		}
		if _, ok := n.Class.(*syntax.Name); !ok {
			return "" // `$y::class`
		}
		return e.classRef(n.Class)
	case *syntax.Literal:
		if n.LitKind != syntax.LitString {
			return ""
		}
		s, ok := plainString(n.Raw)
		s = strings.TrimPrefix(s, `\`)
		if !ok || s == "" || !validClassName(s) {
			return ""
		}
		return s
	}
	return ""
}

// validClassName reports a plain (possibly qualified) class name.
func validClassName(s string) bool {
	for _, part := range strings.Split(s, `\`) {
		if part == "" {
			return false
		}
		for i := 0; i < len(part); i++ {
			c := part[i]
			if !(c == '_' || c >= 0x80 || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (i > 0 && c >= '0' && c <= '9')) {
				return false
			}
		}
	}
	return true
}

// isACond narrows t (the first argument of call c) by `is_a($x, Foo::class
// [, $allowString])` or `is_subclass_of(…)`.
func (e *Env) isACond(t types.Type, c *syntax.FuncCall, fn string, truthy bool) types.Type {
	args := c.Args.Args
	if len(args) < 2 || len(args) > 3 {
		return t
	}
	for _, x := range args {
		if a, ok := x.(*syntax.Arg); !ok || a.Unpack || a.Name != nil {
			return t
		}
	}
	cls := e.classNameValue(args[1].(*syntax.Arg).Value)
	if cls == "" {
		return t
	}
	allowString := fn == "is_subclass_of"
	if len(args) == 3 {
		allowString = constLiteral(args[2].(*syntax.Arg).Value) != "false"
	}
	if !truthy {
		// An object failing the check is no instance (is_a) or no strict
		// subclass (is_subclass_of) of cls; strings are left alone.
		return e.notInstance(t, cls, fn == "is_a")
	}
	out := instanceType(t, cls)
	if allowString && (t.Has("string") || t.Has("mixed")) {
		out = types.Union(out, types.Of("string"))
	}
	return out
}

// inArrayCond narrows t (the needle of c) by a strict `in_array($x, $list,
// true)` that holds: $x has the type of the list's elements.
func (e *Env) inArrayCond(t types.Type, c *syntax.FuncCall, truthy bool) types.Type {
	args := c.Args.Args
	if !truthy || len(args) != 3 {
		return t
	}
	for _, x := range args {
		if a, ok := x.(*syntax.Arg); !ok || a.Unpack || a.Name != nil {
			return t
		}
	}
	if constLiteral(args[2].(*syntax.Arg).Value) != "true" {
		return t
	}
	ht := e.TypeOf(args[1].(*syntax.Arg).Value)
	var el types.Type
	if ht.IsSealedShape() && ht.OnlyOf("array") {
		// A literal list: the union of its values.
		var vs []types.Type
		for _, k := range ht.ShapeKeys() {
			vs = append(vs, k.Type)
		}
		if len(vs) == 0 {
			return t
		}
		el = types.Union(vs...)
	} else {
		for _, a := range ht.Atoms() {
			if !strings.HasSuffix(a, "[]") && a != "null" && a != "false" {
				return t // plain array, iterable, mixed…: elements unknown
			}
		}
		el = ht.Elem()
	}
	if el.IsUnknown() || el.Has("mixed") {
		return t
	}
	return e.assertIs(t, el.WithoutArrayInfo())
}
