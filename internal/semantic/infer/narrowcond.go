package infer

import (
	"math"
	"slices"
	"strconv"
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// Comparison and type-guard narrowing: `===`/`==` and their negations
// (null/true/false literals, scalar literals and constants, `true ===
// cond`, `gettype($x) === 'string'`, `$x::class === Foo::class`), chains
// whose base must be non-null for the condition to hold (`$x?->m()`,
// `isset($x->p)`), `is_a()`, `is_subclass_of()` and strict `in_array()`.

// truthyType is t where it is truthy: neither null nor false, `true` for
// bool, a non-empty array, an int with known values other than 0.
func truthyType(t types.Type) types.Type {
	t = replaceAtom(t.Without("null", "false"), "bool", "true").WithNonEmpty(true).IntsWithout(0)
	if _, known := t.StringLengthRanges(); known {
		t = t.LengthsWithout(0) // '' is falsy ('0' has the length of others)
	}
	return t
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
	t = t.IntsWithin(0, 0) // known int values: only 0 is falsy
	if _, known := t.StringLengthRanges(); known {
		t = t.LengthsWithin(0, 1) // '' or '0'
	}
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
			// Emptied, the path is impossible (as with truthyType).
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

// instanceOf narrows t to the values that are instances of class cls: the
// members of t that are cls or its subtypes, plus cls itself when another
// member may hold an instance of it (a parent class, an interface, an
// unknown class; mixed, object, iterable, callable). Nothing left (`string`
// tested against a class) is an impossible branch: unknown.
func (e *Env) instanceOf(t types.Type, cls string) types.Type {
	if t.HasAny("mixed", "object", "iterable", "callable") {
		return instanceType(t, cls)
	}
	iface := func(c string) bool {
		k := e.Index.Class(c, e.PHP)
		return k == nil || k.Kind == syntax.KindInterface
	}
	var keep []string
	possible := false
	for _, a := range t.Classes() {
		c := strings.TrimPrefix(a, `\`)
		switch {
		case strings.EqualFold(c, cls) || e.Index.IsSubtype(c, cls, e.PHP):
			keep = append(keep, a)
		case e.Index.IsSubtype(cls, c, e.PHP) || iface(c) || iface(cls):
			possible = true
		}
	}
	var parts []types.Type
	for _, a := range keep {
		parts = append(parts, types.Of(a).WithTypeArgs(a, t.TypeArgs(a)))
	}
	if possible {
		parts = append(parts, instanceType(t, cls))
	}
	if len(parts) == 0 {
		return types.Unknown
	}
	return types.Union(parts...)
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
func (e *Env) eqCond(t types.Type, c *syntax.Binary, name string, truthy bool, steps *condBudget) types.Type {
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
	for _, p := range [2][2]syntax.Expr{{l, r}, {r, l}} {
		if !e.lengthOperand(p[0], name) {
			continue
		}
		if v, ok := e.knownInt(p[1]); ok {
			return lengthCmp(t, c.Op.Kind, v, truthy)
		}
	}
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
		searched := p[1]
		if a, ok := searched.(*syntax.Assign); ok && a.Op.Kind == syntax.TEqual && !a.ByRef {
			searched = syntax.UnwrapParens(a.Value) // `false === $i = strpos($x, 'a')`
		}
		if call, ok := searched.(*syntax.FuncCall); ok && lit == "false" && !holds && e.needleFound(call, name) {
			return t.Without("null", "false") // `strpos($x, 'a') !== false`: found
		}
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
	if _, known := t.IntRanges(); known {
		if v, ok := e.knownInt(other); ok {
			return intCmp(t, v, strict, holds)
		}
	}
	if _, known := t.StringLengthRanges(); known {
		if l, ok := syntax.UnwrapParens(other).(*syntax.Literal); ok && l.LitKind == syntax.LitString {
			if text, ok := syntax.LiteralStringValue(l); ok && text == "" {
				return emptyCmp(t, strict, holds)
			}
		}
	}
	lit := constLiteral(other)
	if t.StringsTracked() {
		// `$s === 'image/png'`; loosely only for a non-numeric literal, the
		// one strings compare by identity with.
		if v, ok := stringLiteral(other); ok && (strict || !numericText(v)) {
			if holds {
				t = t.StringsOnly(v)
			} else {
				t = t.StringsWithout(v)
			}
		}
	}
	if !strict {
		switch {
		case lit == "true" || lit == "false":
			// `$x == false` compares $x as a bool: a truthiness test.
			if (lit == "true") == holds {
				return truthyType(t)
			}
			return falsyType(t)
		case !holds:
			// Not loosely equal to null, '', 0 or 0.0: neither null nor
			// false, which equal each of them; not equal to '0': not false.
			return t.Without(looseFalsyPeers(other, lit)...)
		case lit == "" && objectOnly(e.TypeOf(other)):
			// `$x == $object`: null equals no object.
			return t.Without("null")
		}
		return t
	}
	if lit == "" {
		if c, name, ok := e.enumCase(other); ok && t.Has(`\`+c.FQN) {
			// `$suit === Suit::Hearts`: that case, or any other.
			cls := `\` + c.FQN
			if holds {
				return e.assertIs(t, types.Of(cls)).CasesOnly(cls, name)
			}
			return t.CasesWithout(cls, name, enumCaseNames(c))
		}
		// `$x === 'a'`, `$x === Kind::A`: the constant's type.
		if holds {
			if vt, ok := e.constValueType(other); ok {
				if e.native && !slices.ContainsFunc(t.Atoms(), func(x string) bool { return e.atomIn(x, vt) }) {
					// Declarations hold in the native environment: identical
					// to a constant of a type the value cannot have, the
					// branch never runs.
					return types.Unknown
				}
				return e.assertIs(t, vt)
			}
			// `$x === $known`: identical to a value that is never null.
			if ot := e.TypeOf(other); !ot.IsUnknown() && !ot.HasAny("null", "mixed", "void") {
				return t.Without("null")
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

// knownInt returns the integer x holds: a literal (with its sign), a plain
// local whose only reaching definition assigns one and that nothing else may
// have changed (`$failed = -1; if ($pid === $failed)`), a fixed builtin
// constant, or the byte length of a known string.
func (e *Env) knownInt(x syntax.Expr) (int64, bool) {
	if v, ok := integerLiteral(x); ok {
		return v, true
	}
	if c, a := lengthArg(x); a != nil {
		// `strlen('12abc')`, `strlen($s)` with $s = '12abc'.
		if l, ok := syntax.UnwrapParens(a).(*syntax.Variable); ok {
			a = e.onlyValue(l)
		}
		if lit, ok := a.(*syntax.Literal); ok && lit.LitKind == syntax.LitString && e.isBuiltinCall(c) {
			if text, ok := syntax.LiteralStringValue(lit); ok {
				return int64(len(text)), true
			}
		}
		return 0, false
	}
	switch v := syntax.UnwrapParens(x).(type) {
	case *syntax.Variable:
		return integerLiteral(e.onlyValue(v))
	case *syntax.ConstFetch:
		return e.constInt(v) // `PHP_SESSION_ACTIVE`
	}
	return 0, false
}

// onlyValue returns the value of the only plain assignment of local v
// reaching it, unwrapped (nil else).
func (e *Env) onlyValue(v *syntax.Variable) syntax.Expr {
	if v.Name == "" || v.NameExpr != nil {
		return nil
	}
	scope := syntax.EnclosingVariableScope(v)
	sv := e.scopeVars(scope)
	defs := sv.defs[v.Name]
	if len(defs) > maxVarDefs {
		return nil
	}
	fwd, back, _ := e.reaching(e.changing(defs), v, scope)
	if len(fwd) != 1 || len(back) != 0 || fwd[0].asg == nil || e.clobbered(sv, fwd, back, v, scope) {
		return nil
	}
	return syntax.UnwrapParens(fwd[0].asg.Value)
}

// emptyCmp narrows t, whose string member has known lengths, compared with
// ”. Loosely, null and false equal ” too.
func emptyCmp(t types.Type, strict, holds bool) types.Type {
	switch {
	case holds && strict:
		if !t.LengthsWithin(0, 0).Has("string") {
			return types.Unknown // never identical: the branch does not run
		}
		return types.StringLengths(types.IntRange{})
	case holds:
		return t.LengthsWithin(0, 0)
	case strict:
		return t.LengthsWithout(0)
	}
	return t.LengthsWithout(0).Without("null", "false")
}

// lengthOperand reports whether x measures the byte length of variable name:
// strlen($x), or mb_strlen($x, '8bit').
func (e *Env) lengthOperand(x syntax.Expr, name string) bool {
	c, a := lengthArg(x)
	return a != nil && isVar(a, name) && e.isBuiltinCall(c)
}

// lengthArg returns the call and its measured argument when x is written as
// strlen(…) or mb_strlen(…, '8bit') (whichever function it resolves to).
func lengthArg(x syntax.Expr) (*syntax.FuncCall, syntax.Expr) {
	c, ok := syntax.UnwrapParens(x).(*syntax.FuncCall)
	if !ok || c.Args == nil || len(c.Args.Args) == 0 {
		return nil, nil
	}
	nm, ok := c.Name.(*syntax.Name)
	if !ok {
		return nil, nil
	}
	fn := strings.ToLower(strings.TrimPrefix(nm.Value, `\`))
	if fn != "strlen" && fn != "mb_strlen" {
		return nil, nil
	}
	a, ok := c.Args.Args[0].(*syntax.Arg)
	if !ok || a.Unpack || a.Name != nil {
		return nil, nil
	}
	if fn == "strlen" {
		if len(c.Args.Args) != 1 {
			return nil, nil
		}
		return c, a.Value
	}
	if len(c.Args.Args) != 2 {
		return nil, nil
	}
	enc, ok := c.Args.Args[1].(*syntax.Arg)
	l, lit := syntax.UnwrapParens(enc.Value).(*syntax.Literal)
	if !ok || !lit || l.LitKind != syntax.LitString {
		return nil, nil
	}
	if text, _ := syntax.LiteralStringValue(l); !strings.EqualFold(text, "8bit") {
		return nil, nil
	}
	return c, a.Value
}

// lengthCmp narrows t by a comparison of its byte length with v (op an
// equality or relational operator, holding when truthy): the string member
// keeps the lengths satisfying it, from any length. A length that cannot be
// 0 rules out null and false, whose length is 0.
func lengthCmp(t types.Type, op syntax.TokenKind, v int64, truthy bool) types.Type {
	lo, hi := v, v
	switch op {
	case syntax.TIsIdentical, syntax.TIsEqual, syntax.TIsNotIdentical, syntax.TIsNotEqual:
		if (op == syntax.TIsIdentical || op == syntax.TIsEqual) != truthy {
			if v != 0 {
				return t.LengthsWithout(v)
			}
			return t.LengthsWithout(0).Without("null", "false")
		}
	default:
		lo, hi = orderBounds(op, v, truthy)
	}
	r := t.LengthsWithin(lo, hi)
	if lo > 0 || hi < 0 {
		r = r.Without("null", "false")
	}
	return r
}

// intCmp narrows t, whose int member has known values, compared with the
// integer v: identical, only that int; equal, that int beside the other
// members (which may loosely equal it); different, without v. A loose
// difference from 0 also rules out null and false, which equal 0.
func intCmp(t types.Type, v int64, strict, holds bool) types.Type {
	switch {
	case holds && strict:
		if !t.HasInt(v) {
			return types.Unknown // never identical: the branch does not run
		}
		return types.Ints(types.IntRange{Lo: v, Hi: v})
	case holds:
		return t.IntsWithin(v, v)
	case !strict && v == 0:
		return t.IntsWithout(0).Without("null", "false")
	}
	return t.IntsWithout(v)
}

// intOrder narrows t, whose int member has known values, by `x op v` (op a
// relational operator, x the narrowed variable) holding when truthy.
// Members other than int are kept: PHP compares them with v too.
func intOrder(t types.Type, op syntax.TokenKind, v int64, truthy bool) types.Type {
	return t.IntsWithin(orderBounds(op, v, truthy))
}

// mirrorOrder swaps the operands of a relational operator (`0 < $x` is
// `$x > 0`).
var mirrorOrder = map[syntax.TokenKind]syntax.TokenKind{
	syntax.TLess: syntax.TGreater, syntax.TGreater: syntax.TLess,
	syntax.TIsSmallerOrEqual: syntax.TIsGreaterOrEqual, syntax.TIsGreaterOrEqual: syntax.TIsSmallerOrEqual,
}

// orderBounds is the interval of values x satisfying `x op v` (op a
// relational operator) when truthy, its complement otherwise.
func orderBounds(op syntax.TokenKind, v int64, truthy bool) (lo, hi int64) {
	if !truthy {
		op = map[syntax.TokenKind]syntax.TokenKind{
			syntax.TLess: syntax.TIsGreaterOrEqual, syntax.TIsGreaterOrEqual: syntax.TLess,
			syntax.TGreater: syntax.TIsSmallerOrEqual, syntax.TIsSmallerOrEqual: syntax.TGreater,
		}[op]
	}
	switch op {
	case syntax.TLess:
		return math.MinInt64, v - 1 // integerLiteral never reads the minimum
	case syntax.TIsSmallerOrEqual:
		return math.MinInt64, v
	case syntax.TGreater:
		if v == math.MaxInt64 {
			return 1, 0 // nothing is greater
		}
		return v + 1, math.MaxInt64
	}
	return v, math.MaxInt64
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
		if !ok || a.Unpack || a.Name != nil || !isVar(a.Value, name) || !e.isBuiltinCall(n) {
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

// looseFalsyPeers returns the null and false atoms a value loosely equal to
// constant x (lit: its constLiteral) may have: both for null, ”, 0 and 0.0
// (`null == 0`, `false == ”`), false alone for '0' (`null == '0'` is false).
func looseFalsyPeers(x syntax.Expr, lit string) []string {
	if lit == "null" {
		return []string{"null", "false"}
	}
	l, ok := syntax.UnwrapParens(x).(*syntax.Literal)
	if !ok {
		return nil
	}
	switch l.LitKind {
	case syntax.LitString:
		switch v, ok := syntax.LiteralStringValue(l); {
		case ok && v == "":
			return []string{"null", "false"}
		case ok && v == "0":
			return []string{"false"}
		}
	case syntax.LitInt, syntax.LitFloat:
		if f, err := strconv.ParseFloat(strings.ReplaceAll(l.Raw, "_", ""), 64); err == nil && f == 0 {
			return []string{"null", "false"}
		}
	}
	return nil
}

// objectOnly reports a known type made of classes and object only.
func objectOnly(t types.Type) bool {
	if t.IsUnknown() || len(t.Atoms()) == 0 {
		return false
	}
	for _, a := range t.Atoms() {
		if a != "object" && (!strings.HasPrefix(a, `\`) || strings.HasSuffix(a, "[]")) {
			return false
		}
	}
	return true
}

// stringLiteral returns the value of a plain string literal.
func stringLiteral(x syntax.Expr) (string, bool) {
	l, ok := syntax.UnwrapParens(x).(*syntax.Literal)
	if !ok || l.LitKind != syntax.LitString {
		return "", false
	}
	return syntax.LiteralStringValue(l)
}

// numericText reports whether a string may compare numerically: it reads as
// a number once trimmed (inf and nan included, to stay on the safe side).
func numericText(v string) bool {
	_, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
	return err == nil
}
