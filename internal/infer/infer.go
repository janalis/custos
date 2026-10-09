// Package infer infers expression types using the symbol index.
package infer

import (
	"cmp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"custos/internal/index"
	"custos/internal/names"
	"custos/internal/phpdoc"
	"custos/internal/phpver"
	"custos/internal/syntax"
	"custos/internal/types"
)

// Env infers expression types within one file. Not safe for concurrent use.
type Env struct {
	File  *syntax.File
	Names *names.Resolver
	Index *index.Index // project index layered over the stubs
	PHP   phpver.Version

	cache  map[syntax.Expr]types.Type
	scopes map[syntax.Node]*scopeVars
	busy   map[syntax.Expr]bool
	// bases holds the type of variable reads before the element writes
	// reaching them are applied (see baseType).
	bases map[*syntax.Variable]types.Type

	// Return types inferred from bodies (see bodyReturn), by declaration span.
	bodies    map[syntax.Span]types.Type
	bodyBusy  map[syntax.Span]bool
	bodyDepth int
	decls     map[syntax.Span]syntax.Node

	// Index-time return inference (see AnnotateReturns): the index holds
	// only this file over the stubs, and deps records the namespaced
	// function names resolved through the global fallback.
	annotating bool
	deps       []string

	// generics caches genBindings.
	generics map[string]map[string]tplBindings

	// docs caches DocOf; docScopes caches the template names and type
	// aliases in scope at a declaration (see resolverFor).
	docs      map[syntax.Node]*phpdoc.Doc
	docScopes map[syntax.Node]*docScope

	// asserts caches assertsOf.
	asserts map[syntax.Expr]*callAsserts

	// Untyped property inference (see propinfer.go): writes per class,
	// results and recursion guard by property span, classes by span.
	propWritesCache map[*syntax.ClassLike]*classWrites
	props           map[syntax.Span]types.Type
	propBusy        map[syntax.Span]bool
	classes         map[syntax.Span]syntax.Node

	// guardIdx caches guardIndexOf by statement-list owner (nil: file).
	guardIdx map[syntax.Node]*guardIndex

	// conds caches parsedCond; builtinCalls caches isBuiltinCall.
	conds        map[string]*types.Cond
	builtinCalls map[*syntax.FuncCall]bool

	// dynReads caches dynamicRead per variable read (computed by
	// variableBase); assignsMemo caches stmtAssigns.
	dynReads    map[*syntax.Variable]uint8
	assignsMemo map[assignKey]bool

	// native: types come from native declarations only (see Native);
	// nativeTwin caches the native Env of this one.
	native     bool
	nativeTwin *Env
}

// Native returns an Env over the same file and index whose types ignore
// user PHPDoc: parameter, property and return docs, inline @var, templates,
// conditional returns, assertions and @param-out of project declarations
// (builtin stub docs still count, they describe PHP), and the index-time
// inferred types (computed with docs). A rule can thus tell whether a
// type rests on PHPDoc only (UnnecessaryCasting). Cached per Env.
func (e *Env) Native() *Env {
	if e.native {
		return e
	}
	if e.nativeTwin == nil {
		e.nativeTwin = NewEnv(e.File, e.Names, e.Index, e.PHP)
		e.nativeTwin.native = true
	}
	return e.nativeTwin
}

// IsNative reports whether e is a native Env (see Native).
func (e *Env) IsNative() bool { return e.native }

// userDoc reports whether documented facts of a declaration (builtin: from
// the stubs) are ignored by this Env (see Native).
func (e *Env) userDoc(builtin bool) bool { return e.native && !builtin }

// docScope holds the @template names and type aliases declared on a
// declaration and its enclosing ones.
type docScope struct {
	tpl     map[string]bool
	aliases map[string]string
}

// DocOf returns the parsed doc comment of declaration n (nil when it has
// none), parsed once per Env.
func (e *Env) DocOf(n syntax.Node) *phpdoc.Doc {
	if d, ok := e.docs[n]; ok {
		return d
	}
	var d *phpdoc.Doc
	if c := index.DocComment(e.File, n); c != "" {
		d = phpdoc.Parse(c)
	}
	if e.docs == nil {
		e.docs = map[syntax.Node]*phpdoc.Doc{}
	}
	e.docs[n] = d
	return d
}

// scopeDocs returns the template names and aliases declared on the
// function-likes and classes enclosing n (n included), innermost last.
func (e *Env) scopeDocs(n syntax.Node) *docScope {
	for n != nil {
		switch n.(type) {
		case *syntax.Function, *syntax.Method, *syntax.Closure, *syntax.PropertyHook, *syntax.ClassLike:
		default:
			n = n.Parent()
			continue
		}
		break
	}
	if n == nil {
		return nil
	}
	if s, ok := e.docScopes[n]; ok {
		return s
	}
	outer := e.scopeDocs(n.Parent())
	s := outer
	if d := e.DocOf(n); d != nil {
		names, aliases := d.Templates(), d.TypeAliases()
		if len(names) > 0 || len(aliases) > 0 {
			s = &docScope{tpl: map[string]bool{}, aliases: map[string]string{}}
			if outer != nil {
				for k := range outer.tpl {
					s.tpl[k] = true
				}
				for k, v := range outer.aliases {
					s.aliases[k] = v
				}
			}
			for _, t := range names {
				s.tpl[t] = true
			}
			for k, v := range aliases {
				s.aliases[k] = v
			}
		}
	}
	if e.docScopes == nil {
		e.docScopes = map[syntax.Node]*docScope{}
	}
	e.docScopes[n] = s
	return s
}

// NewEnv creates an inference environment.
func NewEnv(f *syntax.File, r *names.Resolver, ix *index.Index, php phpver.Version) *Env {
	return &Env{
		File: f, Names: r, Index: ix, PHP: php, cache: map[syntax.Expr]types.Type{}, scopes: map[syntax.Node]*scopeVars{}, busy: map[syntax.Expr]bool{},
		bodies: map[syntax.Span]types.Type{}, bodyBusy: map[syntax.Span]bool{},
	}
}

func (e *Env) resolver(at uint32) types.Resolver {
	return func(w string) string { return e.className(w, at) }
}

// className resolves a class name written in a doc type at offset at. For
// a "!name" probe (see names.Resolver.Class) a class declared under that
// name in the current namespace also counts.
func (e *Env) className(w string, at uint32) string {
	if probe, ok := strings.CutPrefix(w, "!"); ok {
		if fqn := e.Names.Class(w, at); fqn != "" {
			return fqn
		}
		if fqn := e.Names.Class(probe, at); e.Index.Class(fqn, e.PHP) != nil {
			return fqn
		}
		return ""
	}
	return e.Names.Class(w, at)
}

// resolverFor resolves names in docs attached around n, mapping @template
// names declared on the enclosing function/method/class to "" (mixed).
func (e *Env) resolverFor(n syntax.Node, at uint32) types.Resolver {
	ds := e.scopeDocs(n)
	if ds == nil {
		return e.resolver(at)
	}
	return func(w string) string {
		if ds.tpl[w] {
			return ""
		}
		if def, ok := ds.aliases[w]; ok {
			return "=" + def
		}
		return e.className(w, at)
	}
}

// TypeOf returns the inferred type of x (types.Unknown when it cannot be determined).
func (e *Env) TypeOf(x syntax.Expr) types.Type {
	if x == nil {
		return types.Unknown
	}
	if t, ok := e.cache[x]; ok {
		return t
	}
	if e.busy[x] {
		return types.Unknown // recursion (e.g. $a = $a + 1)
	}
	e.busy[x] = true
	t := e.infer(x)
	delete(e.busy, x)
	e.cache[x] = t
	return t
}

// isFirstClassCallable reports `f(...)`-style argument lists.
func isFirstClassCallable(a *syntax.ArgList) bool {
	if a == nil || len(a.Args) != 1 {
		return false
	}
	_, ok := a.Args[0].(*syntax.VariadicPlaceholder)
	return ok
}

func (e *Env) infer(x syntax.Expr) types.Type {
	switch n := x.(type) {
	case *syntax.FuncCall:
		if isFirstClassCallable(n.Args) {
			return types.Of(`\Closure`)
		}
	case *syntax.MethodCall:
		if isFirstClassCallable(n.Args) {
			return types.Of(`\Closure`)
		}
	case *syntax.StaticCall:
		if isFirstClassCallable(n.Args) {
			return types.Of(`\Closure`)
		}
	}
	switch n := x.(type) {
	case *syntax.Literal:
		switch n.LitKind {
		case syntax.LitInt:
			return types.Int
		case syntax.LitFloat:
			return types.Float
		}
		return types.String
	case *syntax.InterpolatedString:
		if n.Backtick {
			return types.Of("string", "null", "false")
		}
		return types.String
	case *syntax.MagicConst:
		if n.Token.Kind == syntax.TLine {
			return types.Int
		}
		return types.String
	case *syntax.Paren:
		return e.TypeOf(n.Expr)
	case *syntax.ConstFetch:
		return e.constType(n)
	case *syntax.Array:
		return e.arrayType(n)
	case *syntax.Unary:
		return e.unaryType(n)
	case *syntax.Binary:
		return e.binaryType(n)
	case *syntax.Assign:
		if n.Op.Kind == syntax.TEqual {
			return e.TypeOf(n.Value)
		}
		return e.compoundType(n)
	case *syntax.Ternary:
		then := n.Then
		if then == nil {
			return types.Union(e.TypeOf(n.Cond).Without("null", "false"), e.TypeOf(n.Else))
		}
		return types.Union(e.TypeOf(then), e.TypeOf(n.Else))
	case *syntax.Instanceof, *syntax.Isset, *syntax.Empty:
		return types.Bool
	case *syntax.Print:
		return types.Int
	case *syntax.Yield:
		return e.yieldType(n)
	case *syntax.YieldFrom:
		return e.yieldFromType(n)
	case *syntax.Exit:
		return e.exitType(n)
	case *syntax.Throw:
		return types.Of("never")
	case *syntax.Clone:
		return e.cloneType(n)
	case *syntax.Closure, *syntax.ArrowFunction:
		return e.closureType(n)
	case *syntax.IncDec:
		return e.TypeOf(n.Var)
	case *syntax.Match:
		var ts []types.Type
		for _, a := range n.Arms {
			ts = append(ts, e.TypeOf(a.Body))
		}
		if len(ts) == 0 {
			return types.Unknown
		}
		return types.Union(ts...)
	case *syntax.New:
		return e.newType(n)
	case *syntax.Variable:
		return e.variableType(n)
	case *syntax.FuncCall:
		return e.funcCallType(n)
	case *syntax.MethodCall:
		return e.methodCallType(n)
	case *syntax.StaticCall:
		return e.staticCallType(n)
	case *syntax.PropertyFetch:
		t := e.propertyType(e.TypeOf(n.Var), n.Name, false)
		if key := narrowKey(n); key != "" {
			t = e.narrowExpr(t, n, key, syntax.EnclosingVariableScope(n))
		}
		return t
	case *syntax.StaticPropertyFetch:
		cls := e.classRef(n.Class)
		if cls == "" {
			return types.Unknown
		}
		v, ok := n.Name.(*syntax.Variable)
		if !ok || v.Name == "" {
			return types.Unknown
		}
		if p := e.Index.FindProperty(cls, v.Name, e.PHP); p != nil {
			t := e.propType(p, cls)
			if key := narrowKey(n); key != "" {
				t = e.narrowExpr(t, n, key, syntax.EnclosingVariableScope(n))
			}
			return t
		}
		return types.Unknown
	case *syntax.ClassConstFetch:
		return e.classConstType(n)
	case *syntax.ArrayDimFetch:
		t := e.dimType(n)
		if key := narrowKey(n); key != "" && !t.IsUnknown() {
			t = e.narrowExpr(t, n, key, syntax.EnclosingVariableScope(n))
		}
		return t
	}
	return types.Unknown
}

// sealedComputedDim types an element read with a computed key on an empty
// literal array: one of the values written into it since (a missing key
// reads null with a warning, which is not added, as for `T[]` elements). Unknown with an unknown write; ok is false for
// another type, or nothing known.
func (e *Env) sealedComputedDim(ct types.Type, n *syntax.ArrayDimFetch) (types.Type, bool) {
	// (A literal key on a sealed shape is shapeDim's: dimType asks it
	// first.) Only an empty literal: with listed keys, any of their values
	// may be read, a union rules would distrust.
	if !ct.IsSealedShape() || len(ct.ShapeKeys()) > 0 || n.Dim == nil || !ct.Without("null").IsArrayLike() {
		return types.Unknown, false
	}
	el := e.widenElem(types.Of("never"), n) // the values written since
	if el.IsUnknown() || el.OnlyOf("never") {
		return types.Unknown, false
	}
	if v := asVariable(n.Var); v != nil {
		if ws, _ := e.reachingWrites(v); slices.ContainsFunc(ws, func(w *elemWrite) bool { return w.nested }) {
			// `$by[$k][] = $row`: the stored arrays grew since.
			el = el.WithoutArrayInfo()
		}
	}
	return el, true
}

// asVariable returns x as a plain variable (nil otherwise).
func asVariable(x syntax.Expr) *syntax.Variable {
	if v, ok := syntax.UnwrapParens(x).(*syntax.Variable); ok && v.Name != "" && v.Name != "this" {
		return v
	}
	return nil
}

// dimType is the type of element read n before narrowing.
func (e *Env) dimType(n *syntax.ArrayDimFetch) types.Type {
	ct := e.baseType(n.Var)
	if ct.IsUnknown() {
		return types.Unknown
	}
	if t, ok := e.shapeDim(ct, n); ok {
		return t
	}
	if t, ok := e.sealedComputedDim(ct, n); ok {
		return t
	}
	// X[]|null (or |false: a failed builtin) indexes to X; a plain `array`
	// member has unknown elements.
	if el := ct.Elem(); !el.IsUnknown() && ct.Without("null", "false").IsArrayLike() && !ct.Has("array") {
		return e.widenElem(el, n)
	}
	if ct.Without("null", "false").OnlyOf("string") {
		return types.String
	}
	return types.Unknown
}

func (e *Env) constType(n *syntax.ConstFetch) types.Type {
	switch v := strings.ToLower(strings.TrimPrefix(n.Name.Value, `\`)); v {
	case "true", "false":
		return types.Of(v)
	case "null":
		return types.Null
	case "php_int_max", "php_int_min", "php_int_size", "php_version_id", "php_major_version", "php_minor_version", "e_all", "e_error", "e_warning", "e_notice", "e_strict", "e_deprecated":
		return types.Int
	case "php_eol", "php_version", "php_os", "php_os_family", "directory_separator", "path_separator":
		return types.String
	case "php_float_epsilon", "php_float_max", "php_float_min", "m_pi", "nan", "inf":
		return types.Float
	}
	fqn, fb := e.Names.Const(n.Name.Value, n.Span().Start)
	c := e.Index.Constant(fqn, e.PHP)
	if c == nil && fb != "" {
		c = e.Index.Constant(fb, e.PHP)
	}
	if c == nil {
		return types.Unknown
	}
	return literalTextType(c.Value)
}

// literalTextType infers the type of a constant initialiser from its source text.
func literalTextType(v string) types.Type {
	v = strings.TrimSpace(v)
	if v == "" {
		return types.Unknown
	}
	switch c := v[0]; {
	case c == '\'' || c == '"':
		return types.String
	case c == '[' || strings.HasPrefix(strings.ToLower(v), "array("):
		return types.Array
	case (c >= '0' && c <= '9') || c == '-' || c == '.':
		if strings.ContainsAny(v, ".eE") && !strings.HasPrefix(strings.ToLower(v), "0x") {
			return types.Float
		}
		return types.Int
	}
	switch strings.ToLower(v) {
	case "true", "false":
		return types.Bool
	case "null":
		return types.Null
	}
	return types.Unknown
}

func (e *Env) unaryType(n *syntax.Unary) types.Type {
	switch n.Op.Kind {
	case syntax.TExclaim:
		return types.Bool
	case syntax.TIntCast:
		return types.Int
	case syntax.TDoubleCast:
		return types.Float
	case syntax.TStringCast:
		return types.String
	case syntax.TArrayCast:
		return types.Array
	case syntax.TBoolCast:
		return types.Bool
	case syntax.TObjectCast:
		return types.Of("object")
	case syntax.TUnsetCast:
		return types.Null
	case syntax.TVoidCast:
		return types.Void
	case syntax.TTilde:
		// ~ on a string flips its bytes and yields a string.
		switch t := e.TypeOf(n.Expr); {
		case t.OnlyOf("string"):
			return types.String
		case mayBeString(t):
			return types.Unknown
		}
		return types.Int
	case syntax.TAt:
		return e.TypeOf(n.Expr)
	default: // TMinus, TPlus: the parser builds no other unary operator
		t := e.TypeOf(n.Expr)
		if t.OnlyOf("int") || t.OnlyOf("float") {
			return t
		}
		return types.Unknown
	}
}

func (e *Env) numeric(a, b syntax.Expr) types.Type {
	ta, tb := e.TypeOf(a), e.TypeOf(b)
	switch {
	case ta.OnlyOf("int") && tb.OnlyOf("int"):
		return types.Int
	case ta.OnlyOf("int", "float") && tb.OnlyOf("int", "float"):
		return types.Float
	}
	return types.Unknown
}

func (e *Env) binaryType(n *syntax.Binary) types.Type {
	switch n.Op.Kind {
	case syntax.TPipe:
		return e.pipeType(n.Left, n.Right)
	case syntax.TDot:
		return types.String
	case syntax.TPlus:
		if e.TypeOf(n.Left).IsArrayLike() && e.TypeOf(n.Right).IsArrayLike() {
			return types.Array
		}
		return e.numeric(n.Left, n.Right)
	case syntax.TMinus, syntax.TMul, syntax.TPow:
		return e.numeric(n.Left, n.Right)
	case syntax.TDiv:
		if t := e.numeric(n.Left, n.Right); !t.IsUnknown() {
			return types.Of("int", "float")
		}
		return types.Unknown
	case syntax.TAmpersand, syntax.TBar, syntax.TCaret:
		return e.bitwise(n.Left, n.Right)
	case syntax.TMod, syntax.TSl, syntax.TSr, syntax.TSpaceship:
		return types.Int
	case syntax.TCoalesce:
		return types.Union(e.TypeOf(n.Left).Without("null"), e.TypeOf(n.Right))
	default: // All remaining parser binary operators are comparisons or boolean operators.
		return types.Bool
	}
}

func (e *Env) compoundType(n *syntax.Assign) types.Type {
	switch n.Op.Kind {
	case syntax.TConcatEqual:
		return types.String
	case syntax.TPlusEqual, syntax.TMinusEqual, syntax.TMulEqual, syntax.TPowEqual:
		return e.numeric(n.Var, n.Value)
	case syntax.TDivEqual: // as `/`
		if t := e.numeric(n.Var, n.Value); !t.IsUnknown() {
			return types.Of("int", "float")
		}
		return types.Unknown
	case syntax.TCoalesceEqual:
		lt := e.TypeOf(n.Var)
		if d, ok := syntax.UnwrapParens(n.Var).(*syntax.ArrayDimFetch); ok && lt.IsUnknown() {
			// `$by[$k] ??= []` on a literal array no write has filled yet:
			// the key is absent, the result is the new value.
			if v := asVariable(d.Var); v != nil && d.Dim != nil {
				if ct := e.baseType(v); ct.IsSealedShape() && len(ct.ShapeKeys()) == 0 {
					if ws, _ := e.reachingWrites(v); len(ws) == 0 {
						return e.TypeOf(n.Value)
					}
				}
			}
		}
		return types.Union(lt.Without("null"), e.TypeOf(n.Value))
	case syntax.TAndEqual, syntax.TOrEqual, syntax.TXorEqual:
		return e.bitwise(n.Var, n.Value)
	default: // %=, <<=, >>= (syntax.TokenKind.IsAssignOp lists every compound operator)
		return types.Int
	}
}

// bitwise types `&`, `|`, `^`: a string when both operands are strings
// (PHP operates on the bytes), else an int; unknown when both may be strings.
func (e *Env) bitwise(a, b syntax.Expr) types.Type {
	ta, tb := e.TypeOf(a), e.TypeOf(b)
	switch {
	case ta.OnlyOf("string") && tb.OnlyOf("string"):
		return types.String
	case mayBeString(ta) && mayBeString(tb):
		return types.Unknown
	}
	return types.Int
}

// mayBeString reports a type that is unknown or may hold a string.
func mayBeString(t types.Type) bool { return t.IsUnknown() || t.HasAny("string", "mixed") }

// ---- classes -------------------------------------------------------------------------

// ClassFQN returns the FQN of a class declaration ("" for anonymous classes).
func (e *Env) ClassFQN(c *syntax.ClassLike) string { return e.Names.DeclFQN(c) }

// selfClass is the class `$this`, `self`, `static` and `new static` denote
// inside class-like c: its FQN, but "" (unknown) in a trait, whose methods
// run as members of the using class (no object is an instance of a
// trait).
func (e *Env) selfClass(c *syntax.ClassLike) string {
	if c != nil && c.ClassKind == syntax.KindTrait {
		return ""
	}
	return e.ClassFQN(c)
}

// classRef resolves a class reference expression (Name or expression) to a FQN.
func (e *Env) classRef(x syntax.Expr) string {
	switch n := x.(type) {
	case *syntax.Name:
		low := strings.ToLower(n.Value)
		switch low {
		case "self", "static":
			return e.selfClass(syntax.EnclosingClass(n))
		case "parent":
			return e.Names.ParentFQN(syntax.EnclosingClass(n))
		}
		return e.Names.Class(n.Value, n.Span().Start)
	default:
		t := e.TypeOf(x)
		if cs := t.Classes(); len(cs) == 1 {
			return strings.TrimPrefix(cs[0], `\`)
		}
	}
	return ""
}

// anonClassType types an instance of anonymous class cl as the
// intersection of its parent and interfaces (`new class extends P
// implements I {}` is `\P&\I`), a valid declared type whose members are
// found on either side; object without any. Its own extra methods are not
// known (no name could be written for it).
func (e *Env) anonClassType(cl *syntax.ClassLike) types.Type {
	var cs []string
	if p := e.Names.ParentFQN(cl); p != "" {
		cs = append(cs, `\`+p)
	}
	for _, i := range cl.Implements {
		cs = append(cs, `\`+e.Names.Class(i.Value, i.Span().Start))
	}
	if len(cs) == 0 {
		return types.Of("object")
	}
	return types.Intersect(cs...)
}

func (e *Env) newType(n *syntax.New) types.Type {
	if cl, ok := n.Class.(*syntax.ClassLike); ok {
		return e.anonClassType(cl)
	}
	if cls := e.classRef(n.Class); cls != "" {
		return types.Of(`\` + cls)
	}
	return types.Unknown
}

// bindStatic replaces static/self in a member type by the receiver class.
func bindStatic(t types.Type, receiver string) types.Type {
	if t.IsUnknown() || receiver == "" || !t.HasAny("static", "self", "static[]", "self[]") {
		return t
	}
	atoms := make([]string, 0, len(t.Atoms()))
	for _, a := range t.Atoms() {
		switch a {
		case "static", "self":
			atoms = append(atoms, `\`+receiver)
		case "static[]", "self[]":
			atoms = append(atoms, `\`+receiver+"[]")
		default:
			atoms = append(atoms, a)
		}
	}
	return types.Of(atoms...).WithTypeArgsFrom(t)
}

// memberType picks the declared type, falling back to the doc type; when
// both exist the doc type wins if it is more specific (e.g. Foo[] vs array).
func memberType(declared, doc string) types.Type {
	d := types.FromDoc(declared, nil)
	if doc == "" {
		return d
	}
	return pickMemberType(d, types.FromDoc(doc, nil))
}

// builtinMemberType is memberType for a builtin declaration (stubs): the
// declared type is the real, version-resolved signature, so the doc type
// only refines its members (`string[]` for `array`) and never widens it
// (`substr()` documents `string|false`, its 8.0+ signature is `string`).
func builtinMemberType(declared, doc string) types.Type {
	d := types.FromDoc(declared, nil)
	if doc == "" {
		return d
	}
	return pickMemberType(d, refining(d, types.FromDoc(doc, nil)))
}

// refining keeps the members of the doc type dt that belong to the
// declared type d (all of dt when d is unknown or mixed; d when none does).
func refining(d, dt types.Type) types.Type {
	if d.IsUnknown() || d.Has("mixed") || dt.IsUnknown() {
		return dt
	}
	var drop []string
	for _, a := range dt.Atoms() {
		switch {
		case d.Has(a), (a == "true" || a == "false") && d.Has("bool"),
			strings.HasSuffix(a, "[]") && d.HasAny("array", "iterable"),
			strings.HasPrefix(a, `\`) && d.HasAny("object", "iterable", "callable"),
			a == "callable" && d.Has("callable"):
		default:
			drop = append(drop, a)
		}
	}
	if len(drop) == len(dt.Atoms()) {
		return d
	}
	return dt.Without(drop...)
}

// pickMemberType is memberType for parsed types (dt: the doc type).
func pickMemberType(d, dt types.Type) types.Type {
	if d.IsUnknown() || (d.HasAny("array", "iterable") && !dt.IsUnknown()) || d.Has("mixed") || strictSuperset(dt, d) {
		return dt
	}
	return d.WithTypeArgsFrom(dt) // `@var Collection<int, Foo>` on `: Collection`
}

// strictSuperset reports whether doc lists every class of an object
// declaration plus other classes (or templates): an intersection refining
// the declaration (`@return Mock&T` on `: Mock`), whose extra members must
// not be lost.
func strictSuperset(doc, declared types.Type) bool {
	if doc.IsUnknown() || declared.IsUnknown() || len(doc.Atoms()) <= len(declared.Atoms()) {
		return false
	}
	isClass := func(a string) bool { return strings.HasPrefix(a, `\`) && !strings.HasSuffix(a, "[]") }
	for _, a := range declared.Atoms() {
		if !isClass(a) || !doc.Has(a) {
			return false
		}
	}
	// Only class members (or a template, read as mixed) refine an object
	// declaration; `string|false` docs on `: string` stubs do not.
	for _, a := range doc.Atoms() {
		if !isClass(a) && a != "mixed" {
			return false
		}
	}
	return true
}

func (e *Env) propType(p *index.Property, receiver string) types.Type {
	if e.userDoc(p.Builtin) {
		return bindStatic(types.FromDoc(p.Type, nil), receiver)
	}
	if p.Type == "" && p.DocType == "" {
		return bindStatic(e.inferredProp(p), receiver)
	}
	return bindStatic(memberType(p.Type, p.DocType), receiver)
}

func (e *Env) propertyType(recv types.Type, name syntax.Expr, static bool) types.Type {
	id, ok := name.(*syntax.Identifier)
	if !ok || recv.IsUnknown() {
		return types.Unknown
	}
	var ts []types.Type
	for _, cls := range e.memberClasses(recv, func(c string) bool { return e.Index.FindProperty(c, id.Value, e.PHP) != nil }) {
		c := strings.TrimPrefix(cls, `\`)
		p := e.Index.FindProperty(c, id.Value, e.PHP)
		if p == nil {
			return types.Unknown
		}
		ts = append(ts, e.propType(p, c))
	}
	if len(ts) == 0 {
		return types.Unknown
	}
	return types.Union(ts...)
}

// memberClasses lists the classes of receiver type recv whose member a
// lookup must use: each class of a union (all must have it), but for an
// intersection (`A&B`) only the first class that has the member (has),
// none when no side has it.
func (e *Env) memberClasses(recv types.Type, has func(cls string) bool) []string {
	in := recv.Intersection()
	if in == nil {
		return recv.Classes()
	}
	for _, c := range in {
		if has(strings.TrimPrefix(c, `\`)) {
			return []string{c}
		}
	}
	return nil
}

// methodReturn is the return type of method name called on class cls;
// virtual reports a call that may dispatch to an override (see
// methodBodyReturn). Class templates in the documented return type are
// bound from origin (the receiver class, or the calling class for
// `parent::`) and its generic arguments args.
func (e *Env) methodReturn(cls, name string, virtual bool, origin string, args []types.Type, call *syntax.ArgList) types.Type {
	m := e.Index.FindMethod(cls, name, e.PHP)
	if m == nil {
		return types.Unknown
	}
	if e.userDoc(m.Builtin) {
		return bindStatic(types.FromDoc(m.Return, nil), cls)
	}
	if m.Tpl != nil {
		var classB tplBindings
		c := e.Index.Class(m.Class, e.PHP)
		if c != nil && len(c.Templates) > 0 {
			classB = e.genBindings(origin, args)[strings.ToLower(strings.TrimPrefix(m.Class, `\`))]
		}
		if t, ok := e.tplReturn(m.Tpl, m.Params, call, m.Return, classB, c); ok {
			return bindStatic(t, cls)
		}
	}
	if t, ok := e.genMethodReturn(m, origin, args); ok {
		return bindStatic(t, cls)
	}
	if m.CondReturn != "" {
		if t, ok := e.condCall(m.CondReturn, m.Params, call, m.Return, m.DocReturn); ok {
			return bindStatic(t, cls)
		}
	}
	if m.Return == "" && m.DocReturn == "" {
		// An override without return type keeps the contract of the method
		// it overrides (a parent's or interface's `: mixed`); the body is
		// only consulted when no ancestor declares one.
		if pm := e.inheritedSignature(m); pm != nil {
			if pm.Builtin {
				return bindStatic(builtinMemberType(pm.Return, pm.DocReturn), cls)
			}
			return bindStatic(memberType(pm.Return, pm.DocReturn), cls)
		}
		return e.methodBodyReturn(m, virtual)
	}
	if m.Builtin {
		return bindStatic(builtinMemberType(m.Return, m.DocReturn), cls)
	}
	return bindStatic(memberType(m.Return, m.DocReturn), cls)
}

// inheritedSignature returns the nearest method m overrides (in the parents
// and interfaces of its class) that declares or documents a return type;
// nil when none does.
func (e *Env) inheritedSignature(m *index.Method) *index.Method {
	anc := e.Index.Ancestors(strings.TrimPrefix(m.Class, `\`), e.PHP)
	low := strings.ToLower(m.Name)
	for _, c := range anc[min(1, len(anc)):] {
		if pm, ok := c.Methods[low]; ok && !pm.Magic && (pm.Return != "" || pm.DocReturn != "") {
			return pm
		}
	}
	return nil
}

func (e *Env) methodCallType(n *syntax.MethodCall) types.Type {
	id, ok := n.Name.(*syntax.Identifier)
	if !ok {
		return types.Unknown
	}
	recv := e.TypeOf(n.Var)
	if recv.IsUnknown() {
		return types.Unknown
	}
	var ts []types.Type
	for _, cls := range e.memberClasses(recv, func(c string) bool { return e.Index.FindMethod(c, id.Value, e.PHP) != nil }) {
		c := strings.TrimPrefix(cls, `\`)
		t := e.methodReturn(c, id.Value, true, c, recv.TypeArgs(cls), n.Args)
		if t.IsUnknown() {
			return types.Unknown
		}
		ts = append(ts, t)
	}
	if len(ts) == 0 {
		return types.Unknown
	}
	t := types.Union(ts...)
	if n.NullSafe && recv.IsNullable() {
		t = types.Union(t, types.Null)
	}
	return t
}

func (e *Env) staticCallType(n *syntax.StaticCall) types.Type {
	id, ok := n.Name.(*syntax.Identifier)
	if !ok {
		return types.Unknown
	}
	cls := e.classRef(n.Class)
	if cls == "" {
		return types.Unknown
	}
	origin := cls
	if nm, ok := n.Class.(*syntax.Name); ok && strings.EqualFold(nm.Value, "parent") {
		// parent::m(): the calling class's @extends arguments bind the parent.
		if c := e.selfClass(syntax.EnclosingClass(n)); c != "" {
			origin = c
		}
	}
	return e.methodReturn(cls, id.Value, isVirtualClassRef(n.Class), origin, nil, n.Args)
}

func (e *Env) classConstType(n *syntax.ClassConstFetch) types.Type {
	id, ok := n.Name.(*syntax.Identifier)
	if !ok {
		return types.Unknown
	}
	if strings.EqualFold(id.Value, "class") {
		// `Foo::class`, `self::class`, `parent::class`: class-string<Foo>
		// (not `static::class` nor `$obj::class`, which may name a subclass).
		if nm, ok := n.Class.(*syntax.Name); ok && !strings.EqualFold(nm.Value, "static") {
			if cls := e.classRef(nm); cls != "" {
				return types.ClassString(types.Of(`\` + cls))
			}
		}
		return types.String
	}
	cls := e.classRef(n.Class)
	if cls == "" {
		return types.Unknown
	}
	k := e.Index.FindConst(cls, id.Value, e.PHP)
	if k == nil {
		return types.Unknown
	}
	if k.Case {
		return types.Of(`\` + k.Class)
	}
	if k.Type != "" {
		t := types.FromDoc(k.Type, nil)
		if t.HasAny("self", "parent") {
			c := e.Index.Class(k.Class, e.PHP)
			if c == nil || c.Kind == syntax.KindTrait {
				return types.Unknown
			}
			if t.Has("parent") {
				if c.Parent == "" {
					return types.Unknown
				}
				t = types.Union(t.Without("parent"), types.Of(`\`+c.Parent))
			}
		}
		return bindStatic(t, k.Class)
	}
	return literalTextType(k.Value)
}

// ---- functions -------------------------------------------------------------------------

// ResolveFunction finds the index entry for a named call (nil for dynamic calls).
func (e *Env) ResolveFunction(call *syntax.FuncCall) *index.Function {
	name, ok := call.Name.(*syntax.Name)
	if !ok {
		return nil
	}
	fqn, fb := e.Names.Function(name.Value, name.Span().Start)
	f := e.Index.ResolveFunction(fqn, fb, e.PHP)
	if e.annotating && fb != "" && f != nil && !strings.EqualFold(f.FQN, fqn) {
		e.deps = append(e.deps, fqn)
	}
	return f
}

func (e *Env) funcCallType(n *syntax.FuncCall) types.Type {
	name, named := n.Name.(*syntax.Name)
	if !named {
		return e.invokeType(e.TypeOf(n.Name)) // `$f()`, `(fn() => 1)()`
	}
	if t, ok := e.overrideType(n, name); ok {
		return t
	}
	f := e.ResolveFunction(n)
	if f == nil {
		return types.Unknown
	}
	t := e.declCallType(f, n)
	if st, ok := e.safeCallType(f, n, t); ok {
		return st
	}
	if decls := e.Index.FunctionDecls(f.FQN, e.PHP); len(decls) > maxFuncDecls {
		return types.Unknown // hostile: thousands of declarations of one name
	} else if len(decls) > 1 {
		// Declared more than once (a no-op variant loaded instead of the
		// real one): any declaration may be the one that runs.
		ts := []types.Type{t}
		for _, g := range decls {
			if g != f {
				ts = append(ts, e.declCallType(g, n))
			}
		}
		t = types.Union(ts...)
	}
	return t
}

// safeCallType types a call to thecodingmachine/safe's `Safe\X`, which
// wraps the builtin X and throws where X returns false (and, for the preg_
// functions, null): the builtin's type for these arguments without false
// (and null), unless the Safe declaration's own type t is narrower.
func (e *Env) safeCallType(f *index.Function, n *syntax.FuncCall, t types.Type) (types.Type, bool) {
	ns, base, ok := strings.Cut(f.FQN, `\`)
	if !ok || !strings.EqualFold(ns, "safe") || strings.Contains(base, `\`) {
		return types.Unknown, false
	}
	b := e.Index.Function(base, e.PHP)
	if b == nil || !b.Builtin {
		return types.Unknown, false
	}
	bt, ok := e.overrideType(n, &syntax.Name{Value: `\` + base})
	if !ok {
		bt = e.declCallType(b, n)
	}
	drop := []string{"false"}
	if strings.HasPrefix(strings.ToLower(base), "preg_") || (!t.IsUnknown() && !t.Has("null")) {
		drop = append(drop, "null")
	}
	if nt := bt.Without(drop...); len(nt.Atoms()) > 0 {
		bt = nt
	}
	if bt.IsUnknown() {
		return types.Unknown, false
	}
	if f.Return == "" && f.DocReturn == "" {
		return bt, true // the wrapper's body says nothing about its contract
	}
	if !t.IsUnknown() && !t.Has("mixed") {
		for _, a := range bt.Atoms() {
			if !e.atomIn(a, t) {
				return types.Unknown, false // the Safe declaration knows better
			}
		}
	}
	return bt, true
}

// maxFuncDecls caps the declarations of one function a call unions (see
// funcCallType); beyond it the call is unknown.
const maxFuncDecls = 16

// declCallType types call n to function declaration f.
func (e *Env) declCallType(f *index.Function, n *syntax.FuncCall) types.Type {
	if e.userDoc(f.Builtin) {
		return types.FromDoc(f.Return, nil)
	}
	if f.Tpl != nil {
		if t, ok := e.tplReturn(f.Tpl, f.Params, n.Args, f.Return, nil, nil); ok {
			return t
		}
	}
	if f.CondReturn != "" {
		if t, ok := e.condCall(f.CondReturn, f.Params, n.Args, f.Return, f.DocReturn); ok {
			return t
		}
	}
	if f.Return == "" && f.DocReturn == "" {
		return e.BodyReturnType(f)
	}
	if f.Builtin {
		return builtinMemberType(f.Return, f.DocReturn)
	}
	return memberType(f.Return, f.DocReturn)
}

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
					return e.arrayFilterType(a, t, len(n.Args.Args) > 1), true
				case "array_slice":
					return t.WithoutShape().WithNonEmpty(false), true
				}
				return t.WithoutShape(), true
			}
		}
	case "array_map":
		if cb := arg(0); cb != nil && len(n.Args.Args) > 1 {
			if t, ok := e.arrayMapType(cb); ok {
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
		if e.PHP < phpver.PHP80 && !at.IsNonEmptyArray() {
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

// ---- variables ------------------------------------------------------------------------

type varDef struct {
	pos uint32
	end uint32 // end of the defining construct (0 = pos); uses inside it do not see it
	// kill is the span of the block in which this definition is an
	// unconditional statement (`$x = v;` directly in a `{}` block): uses
	// later in that block no longer see earlier definitions. Zero: none.
	kill syntax.Span
	typ  func() types.Type
	doc  bool // inline @var annotation: overrides the next definition
	// asg is the plain `$x = value;` assignment making the definition
	// (nil for other kinds); boolean aliases read it (see aliasCond).
	asg *syntax.Assign
	// docEnd is the end of the annotation's comment (doc defs only).
	docEnd uint32
	// barrier marks an if/elseif/else chain whose every branch assigns the
	// variable (or leaves): uses within kill (after the chain, same block)
	// forget definitions made before it. Barriers carry no type.
	barrier bool
	// w marks an element write (see elemDefs); such entries carry no type.
	w *elemWrite
}

type scopeVars struct {
	defs map[string][]varDef
	// elemWrites lists, per variable, the writes into its elements (see
	// elemWrite), as definitions positioned at the writing assignment.
	elemWrites map[string][]varDef
	// edefs caches elemDefs.
	edefs map[string][]varDef
	// muts lists the operations that may change each variable (lazy, see
	// mutations()).
	muts map[string][]mutation
	// exits lists the exit regions of the scope (lazy, see exitRegions).
	exits     []exitRegion
	exitsDone bool
	// clobbers holds the positions of extract(), one-argument parse_str(),
	// `$$name = …` writes and include/require, which may set any local
	// (see noteDynamic); dynamic is set by any of them.
	clobbers []uint32
	dynamic  bool
}

// widenElem unions the element type el of the array read by n with every
// value directly written into the same variable's elements by a write that
// can reach the read, so `$a = ['k' => 'x']; $a['l'] = [];` does not type
// `$a['l']` as string.
func (e *Env) widenElem(el types.Type, n *syntax.ArrayDimFetch) types.Type {
	v, ok := n.Var.(*syntax.Variable)
	if !ok || v.Name == "" || v.Name == "this" {
		return el
	}
	ws, back := e.reachingWrites(v)
	if k, ok := literalKey(n.Dim); n.Dim != nil && ok {
		ws, back = writesForKey(ws, back, k)
	}
	return e.widenWrites(el, ws, back)
}

// writesForKey keeps the element writes that may store into literal key k:
// writes to k itself, to a computed key, appends when k is an integer, and
// the nested and unknown writes widenWrites judges itself. `$a[1] =
// explode(…)` does not change `$a[0]`.
func writesForKey(ws []*elemWrite, back []bool, k string) ([]*elemWrite, []bool) {
	var kw []*elemWrite
	var kb []bool
	for i, w := range ws {
		if !w.nested && w.a != nil {
			d := w.dim()
			if d == nil && !types.IsIntKey(k) {
				continue // appends add integer keys
			}
			if wk, ok := literalKey(d); d != nil && ok && wk != k {
				continue
			}
		}
		kw = append(kw, w)
		kb = append(kb, back[i])
	}
	return kw, kb
}

// widenVarElem unions el with every value written into the elements of
// variable v by a write reaching v (see widenElem). Nested writes are
// skipped (see widenKey); an unknown write makes the result unknown.
func (e *Env) widenVarElem(el types.Type, v *syntax.Variable) types.Type {
	ws, back := e.reachingWrites(v)
	return e.widenWrites(el, ws, back)
}

// widenWrites is widenVarElem for the reaching writes ws (back: back edge).
func (e *Env) widenWrites(el types.Type, ws []*elemWrite, back []bool) types.Type {
	if len(ws) == 0 {
		return el
	}
	ts := []types.Type{el}
	for i, w := range ws {
		if w.nested {
			continue // changes an element already counted; see widenKey
		}
		if w.a == nil {
			return types.Unknown
		}
		t := e.writtenType(w.a)
		if t.IsUnknown() && back[i] {
			continue // a back-edge cycle adds nothing (as for variables)
		}
		ts = append(ts, t)
	}
	return types.Union(ts...)
}

// writtenType is the value an element write stores.
func (e *Env) writtenType(a *syntax.Assign) types.Type {
	if a.Op.Kind == syntax.TEqual || a.Op.Kind == syntax.TCoalesceEqual {
		return e.TypeOf(a.Value)
	}
	return e.TypeOf(a)
}

// variableType is the type of a variable read: its reaching definitions
// (variableBase) with the element writes reaching the read applied to its
// array members (withElemWrites), so the value is right wherever it flows.
func (e *Env) variableType(v *syntax.Variable) types.Type {
	if v.Name == "" || v.Name == "this" {
		return e.variableBase(v)
	}
	t, known := e.bases[v] // computed by baseType
	if !known {
		t = e.variableBase(v)
	}
	w, changed := e.withElemWrites(t, v)
	if changed && !known {
		e.setBase(v, t)
	}
	return w
}

func (e *Env) setBase(v *syntax.Variable, t types.Type) {
	if e.bases == nil {
		e.bases = map[*syntax.Variable]types.Type{}
	}
	e.bases[v] = t
}

// baseType is TypeOf, except that a variable read gives its type before
// the element writes reaching it: for the readers that apply those writes
// themselves, key by key (dimType, foreachElem, foreachKey, pointer
// functions), keeping shapes precise.
func (e *Env) baseType(x syntax.Expr) types.Type {
	v, ok := syntax.UnwrapParens(x).(*syntax.Variable)
	if !ok || v.Name == "" || v.Name == "this" {
		return e.TypeOf(x)
	}
	if t, ok := e.bases[v]; ok {
		return t
	}
	if t, ok := e.cache[v]; ok {
		return t // typed already, and no write changed it
	}
	if e.busy[v] {
		return types.Unknown // recursion, as TypeOf
	}
	e.busy[v] = true
	t := e.variableBase(v)
	delete(e.busy, v)
	e.setBase(v, t)
	return t
}

// withElemWrites applies the element writes reaching variable read v to
// the array members of its type t: their element type gains the written
// values (unknown, as `array`, after a nested or unknown write) and shapes
// are dropped. `$a = ['x' => 1]; $a['y'] = 'a'; $b = $a;` types $b as
// (int|string)[], not as the sealed shape {x: int}. Non-array members
// (strings, ArrayAccess objects) are unchanged.
// changed is false when t is returned as is.
func (e *Env) withElemWrites(t types.Type, v *syntax.Variable) (types.Type, bool) {
	if t.IsUnknown() {
		return t, false
	}
	var arr, other []string
	for _, a := range t.Atoms() {
		if a == "array" || strings.HasSuffix(a, "[]") {
			arr = append(arr, a)
		} else {
			other = append(other, a)
		}
	}
	// A write into null creates an array (`?array $n; $n['k'] = 1;`).
	nullOnly := len(other) == 1 && other[0] == "null"
	if len(arr) == 0 && !nullOnly {
		return t, false
	}
	ws, back := e.reachingWrites(v)
	if len(ws) == 0 {
		return t, false
	}
	if nullOnly && e.writeDominates(v) {
		other = nil // every path to v writes into the array: no longer null
	}
	el := t.Elem()
	switch {
	case len(arr) == 0:
		el = types.Of("never") // null only: the writes fill a new array
	case t.IsSealedShape():
		ts := []types.Type{types.Of("never")}
		if !el.IsUnknown() {
			ts = append(ts, el)
		}
		for _, k := range t.ShapeKeys() {
			ts = append(ts, k.Type)
		}
		el = types.Union(ts...)
	case slices.Contains(arr, "array"):
		el = types.Unknown // elements of a plain array are unknown
	}
	for _, w := range ws {
		if w.nested {
			el = types.Unknown // changes an element in place
		}
	}
	if !el.IsUnknown() {
		el = e.widenWrites(el, ws, back)
	}
	if el.IsUnknown() || el.OnlyOf("never") {
		arr = []string{"array"}
	} else {
		arr = arr[:0]
		for _, a := range el.Atoms() {
			arr = append(arr, a+"[]")
		}
	}
	return types.Of(append(other, arr...)...).WithNonEmpty(t.IsNonEmptyArray()), true
}

func (e *Env) variableBase(v *syntax.Variable) types.Type {
	if v.Name == "" {
		return types.Unknown
	}
	if v.Name == "this" {
		if fqn := e.selfClass(syntax.EnclosingClass(v)); fqn != "" {
			return types.Of(`\` + fqn)
		}
		return types.Unknown
	}
	scope := syntax.EnclosingVariableScope(v)
	sv := e.scopeVars(scope)
	defs := sv.defs[v.Name]
	if _, ok := scope.(*syntax.ArrowFunction); ok && len(defs) == 0 {
		// Arrow functions capture the enclosing scope by value.
		outer := e.scopeVars(syntax.EnclosingVariableScope(scope))
		defs = outer.defs[v.Name]
	}
	if len(defs) > maxVarDefs {
		return types.Unknown
	}
	fwd, back, from := e.reaching(defs, v, scope)
	clob := e.clobbered(sv, fwd, back, v, scope)
	undef := !clob && len(fwd)+len(back) > 0 && e.maybeUndefined(sv, defs, fwd, v, scope)
	e.noteDynRead(v, clob, undef)
	if clob {
		return types.Unknown
	}
	ts := make([]types.Type, 0, len(fwd)+len(back))
	after := uint32(0)
	for _, d := range fwd {
		t := d.typ()
		if !t.IsUnknown() && e.forStep(d, v, scope) {
			// The step of a for loop around v runs before the condition
			// is tested again: a back edge, not a definition after it.
			if nt := e.loopCondNarrow(t, d, v, scope); !nt.IsUnknown() {
				ts = append(ts, nt)
			}
			continue
		}
		ts = append(ts, t)
		after = max(after, d.pos, d.end)
	}
	for _, d := range back {
		// An unknown back-edge type (often a cycle through this very use)
		// adds nothing: keep what the forward definitions say.
		if t := d.typ(); !t.IsUnknown() {
			ts = append(ts, e.loopCondNarrow(t, d, v, scope))
		}
	}
	if len(ts) == 0 {
		return types.Unknown
	}
	if undef {
		ts = append(ts, types.Null) // read before any assignment on some path
	}
	t := types.Union(ts...)
	if t.HasShape() || t.IsNonEmptyArray() {
		ms := scope
		if _, ok := scope.(*syntax.ArrowFunction); ok && len(sv.defs[v.Name]) == 0 {
			ms = syntax.EnclosingVariableScope(scope)
		}
		if t.HasShape() && e.shapeClobbered(ms, v.Name) {
			t = t.WithoutShape()
		}
		if t.IsNonEmptyArray() && e.nonEmptyBroken(ms, v.Name, from, v) {
			t = t.WithNonEmpty(false)
		}
	}
	return e.narrow(t, v, scope, after)
}

// forStep reports whether definition d sits in the step expressions of a
// for loop whose body holds v.
func (e *Env) forStep(d varDef, v *syntax.Variable, scope syntax.Node) bool {
	var child syntax.Node = v
	for p := v.Parent(); p != nil && p != scope; child, p = p, p.Parent() {
		if f, ok := p.(*syntax.For); ok && child == syntax.Node(f.Body) {
			for _, x := range f.Loop {
				if sp := x.Span(); d.pos >= sp.Start && d.pos < sp.End {
					return true
				}
			}
		}
	}
	return false
}

// loopCondNarrow narrows the type t of back-edge definition d of v when d
// sits in the condition (or step) of a loop around v that is tested before
// the body runs again: `do { … } while ($e = $e->getPrevious());` only
// loops with a truthy $e; `for (…; $x !== null; $x = next($x))` likewise.
func (e *Env) loopCondNarrow(t types.Type, d varDef, v *syntax.Variable, scope syntax.Node) types.Type {
	in := func(x syntax.Node) bool {
		sp := x.Span()
		return d.pos >= sp.Start && d.pos < sp.End
	}
	for p := v.Parent(); p != nil && p != scope; p = p.Parent() {
		switch n := p.(type) {
		case *syntax.DoWhile:
			if in(n.Cond) {
				return e.applyCond(t, n.Cond, v.Name, true)
			}
		case *syntax.For:
			if len(n.Cond) == 0 {
				continue
			}
			for _, x := range append(slices.Clone(n.Loop), n.Cond...) {
				if in(x) {
					return e.applyCond(t, n.Cond[len(n.Cond)-1], v.Name, true)
				}
			}
		}
	}
	return t
}

// iterElem is the element type of iterating over t: unknown as soon as one
// member (plain `array`, `iterable`, a Traversable class…) has unknown
// elements, so `string|iterable` does not iterate as `string`.
func iterElem(t types.Type) types.Type {
	for _, a := range t.Without("null", "false").Atoms() {
		if !strings.HasSuffix(a, "[]") {
			return types.Unknown
		}
	}
	return t.Elem()
}

// outermostLoop returns the outermost loop statement enclosing n within
// scope, or nil.
func outermostLoop(n, scope syntax.Node) syntax.Node {
	var loop syntax.Node
	for p := n.Parent(); p != nil && p != scope; p = p.Parent() {
		switch p.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			loop = p
		}
	}
	return loop
}

// scopeVars collects variable definitions (assignments, params, foreach,
// catch, inline @var) of one scope, in source order.
func (e *Env) scopeVars(scope syntax.Node) *scopeVars {
	if sv, ok := e.scopes[scope]; ok {
		return sv
	}
	sv := &scopeVars{defs: map[string][]varDef{}, elemWrites: map[string][]varDef{}}
	e.scopes[scope] = sv
	add := func(name string, pos uint32, t func() types.Type) {
		if name != "" {
			sv.defs[name] = append(sv.defs[name], varDef{pos: pos, typ: t})
		}
	}
	var params []*syntax.Param
	var body []syntax.Node
	switch s := scope.(type) {
	case *syntax.Function:
		params, body = s.Params, []syntax.Node{s.Body}
	case *syntax.PropertyHook:
		params = s.Params
		if s.Body != nil {
			body = []syntax.Node{s.Body}
		}
		if strings.EqualFold(s.Name.Value, "set") && len(s.Params) == 0 {
			add("value", s.Span().Start, func() types.Type { return e.hookValueType(s) })
		}
	case *syntax.Method:
		params = s.Params
		if s.Body != nil {
			body = []syntax.Node{s.Body}
		}
	case *syntax.Closure:
		params, body = s.Params, []syntax.Node{s.Body}
		for _, u := range s.Uses {
			add(u.Var.Name, u.Span().Start, func() types.Type {
				outer := e.scopeVars(syntax.EnclosingVariableScope(scope))
				var ts []types.Type
				for _, d := range outer.defs[u.Var.Name] {
					if d.pos < scope.Span().Start && !d.barrier {
						ts = append(ts, d.typ())
					}
				}
				if len(ts) == 0 {
					return types.Unknown
				}
				return types.Union(ts...)
			})
		}
	case *syntax.ArrowFunction:
		params, body = s.Params, []syntax.Node{s.Expr}
	case nil:
		for _, st := range e.File.Stmts {
			body = append(body, st)
		}
	}
	for _, p := range params {
		add(p.Var.Name, p.Span().Start, func() types.Type { return e.paramType(scope, p) })
	}
	// body holds no nil node: the parser always builds function and closure
	// bodies (parseBlock) and arrow function expressions (BadExpr at worst);
	// an abstract method's missing body is not added.
	for _, b := range body {
		syntax.Inspect(b, func(n syntax.Node) bool {
			e.noteDynamic(n, sv)
			switch n := n.(type) {
			case *syntax.Closure:
				if n != scope {
					// A closure importing a variable by reference may write it
					// whenever it is called: its type is unknown from here on.
					for _, u := range n.Uses {
						if u.ByRef && u.Var != nil {
							add(u.Var.Name, n.Span().End, func() types.Type { return types.Unknown })
						}
					}
				}
				return n == scope
			case *syntax.Function, *syntax.Method, *syntax.ArrowFunction, *syntax.PropertyHook, *syntax.ClassLike:
				return n == scope // do not descend into nested scopes
			case *syntax.Assign:
				e.collectDimWrites(n, n, n.Var, sv)
				end := n.Span().End
				var kill syntax.Span
				// Any assignment replaces the value (`.=` is a string, `+=` a
				// number), not only `=` (by reference: not a kill).
				if _, plain := n.Var.(*syntax.Variable); plain && !n.ByRef {
					if es, ok := n.Parent().(*syntax.ExprStmt); ok {
						if blk, ok := es.Parent().(*syntax.Block); ok {
							kill = blk.Span()
						}
					}
				}
				var asg *syntax.Assign
				if v, plain := n.Var.(*syntax.Variable); plain && n.Op.Kind == syntax.TEqual && !n.ByRef && v.NameExpr == nil {
					asg = n
				}
				e.collectAssignTargets(n, func(name string, pos uint32, t func() types.Type) {
					if name != "" {
						sv.defs[name] = append(sv.defs[name], varDef{pos: pos, end: end, kill: kill, typ: t, asg: asg})
					}
				})
			case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall, *syntax.New:
				e.collectOutArgs(n.(syntax.Expr), sv)
			case *syntax.If:
				for _, name := range branchAssigned(n) {
					sv.defs[name] = append(sv.defs[name], varDef{pos: n.Span().Start, barrier: true, kill: joinSpan(n, e.File)})
				}
			case *syntax.Foreach:
				it := n.Expr
				// The loop binds its targets at the start of every iteration:
				// inside the body they hide earlier definitions.
				var kill syntax.Span
				if n.Body != nil {
					kill = n.Body.Span()
				}
				addKill := func(name string, pos uint32, t func() types.Type) {
					if name != "" {
						sv.defs[name] = append(sv.defs[name], varDef{pos: pos, kill: kill, typ: t})
					}
				}
				if n.Key != nil {
					if kv, ok := n.Key.(*syntax.Variable); ok {
						addKill(kv.Name, n.Key.Span().Start, func() types.Type { return e.foreachKey(it) })
					}
				}
				e.collectTargets(n.Value, n.Value.Span().Start, func() types.Type { return e.foreachElem(it) }, addKill)
			case *syntax.Catch:
				if n.Var != nil {
					var atoms []string
					for _, t := range n.Types {
						atoms = append(atoms, `\`+e.Names.Class(t.Value, t.Span().Start))
					}
					add(n.Var.Name, n.Var.Span().Start, func() types.Type { return types.Of(atoms...) })
				}
			case *syntax.Global, *syntax.StaticStmt:
				// types.Unknown provenance.
				syntax.Inspect(n, func(m syntax.Node) bool {
					if v, ok := m.(*syntax.Variable); ok {
						add(v.Name, v.Span().Start, func() types.Type { return types.Unknown })
					}
					return true
				})
				return false
			}
			return true
		})
	}
	// Inline `/** @var Type $name */` comments.
	if !e.native {
		e.inlineVarDocs(scope, func(name string, pos, end uint32, t func() types.Type) {
			sv.defs[name] = append(sv.defs[name], varDef{pos: pos, docEnd: end, typ: t, doc: true})
		})
	}
	for name, defs := range sv.defs {
		sortDefs(defs)
		sv.defs[name] = defs
	}
	return sv
}

func sortDefs(d []varDef) {
	slices.SortStableFunc(d, func(a, b varDef) int { return cmp.Compare(a.pos, b.pos) })
}

// maxVarDefs caps the definitions (assignments, element writes, inline
// annotations) of one variable in one scope that reads consider; beyond it
// the variable is unknown (element writes: an unknown write). Each read
// walks them, so thousands of assignments to one variable with as many
// reads were quadratic; real code stays far below (same cap as value
// discovery in analysis/util).
const maxVarDefs = 512

// collectTargets registers variables written by an assignment target
// (plain variable or list/array destructuring).
func (e *Env) collectTargets(target syntax.Expr, pos uint32, t func() types.Type, add func(string, uint32, func() types.Type)) {
	switch v := target.(type) {
	case *syntax.Variable:
		add(v.Name, pos, t)
	case *syntax.List:
		e.collectItemTargets(v.Items, pos, t, add)
	case *syntax.Array:
		e.collectItemTargets(v.Items, pos, t, add)
	}
}

// collectAssignTargets registers the targets of assignment n. Destructuring
// a variable (`[$a, $b] = $v`) reads its keys as `$v[0]`, `$v[1]` would:
// from its base type, with the element writes reaching n applied key by key.
func (e *Env) collectAssignTargets(n *syntax.Assign, add func(string, uint32, func() types.Type)) {
	t := func() types.Type { return e.TypeOf(n) }
	var items []*syntax.ArrayItem
	switch l := n.Var.(type) {
	case *syntax.List:
		items = l.Items
	case *syntax.Array:
		items = l.Items
	}
	v, ok := syntax.UnwrapParens(n.Value).(*syntax.Variable)
	if items == nil || !ok || v.Name == "" || v.Name == "this" || n.Op.Kind != syntax.TEqual {
		e.collectTargets(n.Var, n.Span().Start, t, add)
		return
	}
	e.collectItems(items, n.Span().Start, func(key string) func() types.Type {
		return func() types.Type {
			ct := e.baseType(v)
			if !ct.Without("null").IsArrayLike() {
				return types.Unknown
			}
			kt, _ := e.shapeKeyOf(ct, key, v)
			return kt
		}
	}, add)
}

// collectItemTargets registers destructuring items; an item reads the
// destructured value's shape key (explicit literal key or position).
func (e *Env) collectItemTargets(items []*syntax.ArrayItem, pos uint32, t func() types.Type, add func(string, uint32, func() types.Type)) {
	e.collectItems(items, pos, func(key string) func() types.Type { return shapeTarget(t, key) }, add)
}

// collectItems registers destructuring items, typing the item of a
// literal key (explicit or positional) with keyType.
func (e *Env) collectItems(items []*syntax.ArrayItem, pos uint32, keyType func(string) func() types.Type, add func(string, uint32, func() types.Type)) {
	for i, it := range items {
		if it == nil || it.Value == nil {
			continue
		}
		key, ok := strconv.Itoa(i), it.Key == nil
		if it.Key != nil {
			key, ok = literalKey(it.Key)
		}
		item := func() types.Type { return types.Unknown }
		if ok && !it.ByRef {
			item = keyType(key)
		}
		e.collectTargets(it.Value, pos, item, add)
	}
}

// collectDimWrites records the element writes `$x[...] = v` made by target,
// part of assignment outer (destructuring targets are recorded as unknown
// writes: a is nil).
func (e *Env) collectDimWrites(outer, a *syntax.Assign, target syntax.Expr, sv *scopeVars) {
	add := func(name string, w *elemWrite) {
		sv.elemWrites[name] = append(sv.elemWrites[name], varDef{pos: outer.Span().Start, end: outer.Span().End, w: w})
	}
	switch t := target.(type) {
	case *syntax.ArrayDimFetch:
		if v, ok := t.Var.(*syntax.Variable); ok && v.Name != "" {
			add(v.Name, &elemWrite{a: a})
			return
		}
		// Nested write: remember which first-level key it changes.
		inner := t
		for {
			d, ok := inner.Var.(*syntax.ArrayDimFetch)
			if !ok {
				break
			}
			inner = d
		}
		if v, ok := inner.Var.(*syntax.Variable); ok && v.Name != "" {
			add(v.Name, &elemWrite{a: a, nested: true, key: inner.Dim})
		}
	case *syntax.List:
		for _, it := range t.Items {
			if it != nil && it.Value != nil {
				e.collectDimWrites(outer, nil, it.Value, sv)
			}
		}
	case *syntax.Array:
		for _, it := range t.Items {
			if it != nil && it.Value != nil {
				e.collectDimWrites(outer, nil, it.Value, sv)
			}
		}
	}
}

func (e *Env) paramType(scope syntax.Node, p *syntax.Param) types.Type {
	at := p.Span().Start
	declared := types.FromNode(p.Type, e.resolver(at))
	if p.Variadic && !declared.IsUnknown() {
		atoms := make([]string, 0, len(declared.Atoms()))
		for _, a := range declared.Atoms() {
			atoms = append(atoms, a+"[]")
		}
		declared = types.Of(atoms...)
	}
	if lit, ok := p.Default.(*syntax.ConstFetch); ok && strings.EqualFold(lit.Name.Value, "null") && !declared.IsUnknown() {
		declared = types.Union(declared, types.Null)
	}
	doc := ""
	if d := e.DocOf(scope); d != nil && !e.native {
		for _, dp := range d.Params() {
			if dp.Name == p.Var.Name {
				doc = dp.Type
			}
		}
	}
	if doc != "" {
		dt := types.FromDoc(doc, e.resolverFor(scope, at))
		if declared.IsUnknown() || declared.HasAny("array", "iterable", "mixed") {
			return dt
		}
		return declared.WithTypeArgsFrom(dt)
	}
	return declared
}

func (e *Env) inlineVarDocs(scope syntax.Node, addDoc func(string, uint32, uint32, func() types.Type)) {
	var span syntax.Span
	if scope == nil {
		span = syntax.Span{Start: 0, End: uint32(len(e.File.Src))}
	} else {
		span = scope.Span()
	}
	toks := e.File.Tokens
	first := sort.Search(len(toks), func(i int) bool { return toks[i].Start >= span.Start })
	excluded := promotedHookDocSpans(scope)
	hook := 0
	for _, t := range toks[first:] {
		if t.Start >= span.End {
			break
		}
		for hook < len(excluded) && excluded[hook].End <= t.Start {
			hook++
		}
		if hook < len(excluded) && excluded[hook].Start <= t.Start {
			continue
		}
		if t.Kind != syntax.TDocComment && t.Kind != syntax.TComment {
			continue
		}
		text := string(e.File.Src[t.Start:t.End])
		if !strings.Contains(text, "@var") {
			continue
		}
		d := phpdoc.Parse(text)
		for _, tag := range d.All("var") {
			typ, rest := phpdoc.SplitType(tag.Text)
			name := phpdoc.VarName(rest)
			if strings.HasPrefix(typ, "$") {
				name = phpdoc.VarName(typ)
				typ, _ = phpdoc.SplitType(rest)
			}
			if name == "" {
				continue
			}
			ty := types.FromDoc(typ, e.resolverFor(e.nodeAt(scope, t.Start), t.Start))
			addDoc(name, t.Start, t.End, func() types.Type { return ty })
		}
	}
}

// semicolonBetween reports whether a `;` token lies in [from, to).
func (e *Env) semicolonBetween(from, to uint32) bool {
	toks := e.File.Tokens
	i := sort.Search(len(toks), func(i int) bool { return toks[i].Start >= from })
	for ; i < len(toks) && toks[i].Start < to; i++ {
		if toks[i].Kind == syntax.TSemicolon {
			return true
		}
	}
	return false
}

// nodeAt returns scope itself (or nil for file scope); used to resolve doc
// names relative to the enclosing declarations.
func (e *Env) nodeAt(scope syntax.Node, _ uint32) syntax.Node { return scope }

// branchAssigned returns the variables assigned (plain `$x = …;` at the top
// level of the branch) in every branch of a complete if/elseif/else chain;
// branches that always leave (return/throw/…) do not constrain the result.
func branchAssigned(n *syntax.If) []string {
	if n.Else == nil {
		return nil
	}
	bodies := []syntax.Stmt{n.Body}
	for _, ei := range n.ElseIfs {
		bodies = append(bodies, ei.Body)
	}
	bodies = append(bodies, n.Else.Body)
	var common map[string]bool
	assignedSomewhere := false
	for _, b := range bodies {
		if terminates(b) {
			continue
		}
		set := topLevelAssigns(b)
		if common == nil {
			common = set
		} else {
			for k := range common {
				if !set[k] {
					delete(common, k)
				}
			}
		}
		assignedSomewhere = true
	}
	if !assignedSomewhere {
		return nil
	}
	out := make([]string, 0, len(common))
	for k := range common {
		out = append(out, k)
	}
	return out
}

func topLevelAssigns(s syntax.Stmt) map[string]bool {
	set := map[string]bool{}
	stmts := []syntax.Stmt{s}
	if b, ok := s.(*syntax.Block); ok {
		stmts = b.Stmts
	}
	for _, st := range stmts {
		es, ok := st.(*syntax.ExprStmt)
		if !ok {
			continue
		}
		if a, ok := es.Expr.(*syntax.Assign); ok && a.Op.Kind == syntax.TEqual {
			if v, ok := a.Var.(*syntax.Variable); ok && v.Name != "" {
				set[v.Name] = true
			}
		}
	}
	return set
}

// joinSpan is the region after an if-chain, up to the end of the enclosing
// statement list.
func joinSpan(n *syntax.If, f *syntax.File) syntax.Span {
	end := uint32(len(f.Src))
	if p := n.Parent(); p != nil {
		end = p.Span().End
	}
	return syntax.Span{Start: n.Span().End, End: end}
}
