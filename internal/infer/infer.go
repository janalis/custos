// Package infer infers expression types using the symbol index.
package infer

import (
	"cmp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"custos/internal/types"

	"custos/internal/index"
	"custos/internal/names"
	"custos/internal/phpdoc"
	"custos/internal/phpver"
	"custos/internal/syntax"
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
}

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
		case *syntax.Function, *syntax.Method, *syntax.Closure, *syntax.ClassLike:
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
	return &Env{File: f, Names: r, Index: ix, PHP: php, cache: map[syntax.Expr]types.Type{}, scopes: map[syntax.Node]*scopeVars{}, busy: map[syntax.Expr]bool{},
		bodies: map[syntax.Span]types.Type{}, bodyBusy: map[syntax.Span]bool{}}
}

func (e *Env) resolver(at uint32) types.Resolver {
	return func(w string) string { return e.Names.Class(w, at) }
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
		return e.Names.Class(w, at)
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
	case *syntax.Exit, *syntax.Throw:
		return types.Of("never")
	case *syntax.Clone:
		return e.TypeOf(n.Expr)
	case *syntax.Closure, *syntax.ArrowFunction:
		return types.Of(`\Closure`)
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
			t = e.narrowExpr(t, n, key, scopeOf(n))
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
			return e.propType(p, cls)
		}
		return types.Unknown
	case *syntax.ClassConstFetch:
		return e.classConstType(n)
	case *syntax.ArrayDimFetch:
		t := e.dimType(n)
		if key := narrowKey(n); key != "" && !t.IsUnknown() {
			t = e.narrowExpr(t, n, key, scopeOf(n))
		}
		return t
	}
	return types.Unknown
}

// dimType is the type of element read n before narrowing.
func (e *Env) dimType(n *syntax.ArrayDimFetch) types.Type {
	ct := e.TypeOf(n.Var)
	if ct.IsUnknown() {
		return types.Unknown
	}
	if t, ok := e.shapeDim(ct, n); ok {
		return t
	}
	// X[]|null indexes to X; a plain `array` member has unknown elements.
	if el := ct.Elem(); !el.IsUnknown() && ct.Without("null").IsArrayLike() && !ct.Has("array") {
		return e.widenElem(el, n)
	}
	if ct.OnlyOf("string") {
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
		return types.Int
	case syntax.TAt:
		return e.TypeOf(n.Expr)
	case syntax.TMinus, syntax.TPlus:
		t := e.TypeOf(n.Expr)
		if t.OnlyOf("int") || t.OnlyOf("float") {
			return t
		}
		return types.Unknown
	}
	return types.Unknown
}

func (e *Env) numeric(a, b syntax.Expr) types.Type {
	ta, tb := e.TypeOf(a), e.TypeOf(b)
	switch {
	case ta.OnlyOf("int") && tb.OnlyOf("int"):
		return types.Int
	case (ta.OnlyOf("int", "float")) && (tb.OnlyOf("int", "float")):
		return types.Float
	}
	return types.Unknown
}

func (e *Env) binaryType(n *syntax.Binary) types.Type {
	switch n.Op.Kind {
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
	case syntax.TMod, syntax.TSl, syntax.TSr, syntax.TAmpersand, syntax.TBar, syntax.TCaret, syntax.TSpaceship:
		return types.Int
	case syntax.TIsEqual, syntax.TIsNotEqual, syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TLess,
		syntax.TIsSmallerOrEqual, syntax.TGreater, syntax.TIsGreaterOrEqual, syntax.TBooleanAnd,
		syntax.TBooleanOr, syntax.TAnd, syntax.TOr, syntax.TXor:
		return types.Bool
	case syntax.TCoalesce:
		return types.Union(e.TypeOf(n.Left).Without("null"), e.TypeOf(n.Right))
	}
	return types.Unknown
}

func (e *Env) compoundType(n *syntax.Assign) types.Type {
	switch n.Op.Kind {
	case syntax.TConcatEqual:
		return types.String
	case syntax.TPlusEqual, syntax.TMinusEqual, syntax.TMulEqual, syntax.TPowEqual:
		return e.numeric(n.Var, n.Value)
	case syntax.TCoalesceEqual:
		return types.Union(e.TypeOf(n.Var).Without("null"), e.TypeOf(n.Value))
	case syntax.TModEqual, syntax.TSlEqual, syntax.TSrEqual, syntax.TAndEqual, syntax.TOrEqual, syntax.TXorEqual:
		return types.Int
	}
	return types.Unknown
}

// ---- classes -------------------------------------------------------------------------

// EnclosingClass returns the class-like declaration containing n (nil outside).
func EnclosingClass(n syntax.Node) *syntax.ClassLike {
	for p := n.Parent(); p != nil; p = p.Parent() {
		if c, ok := p.(*syntax.ClassLike); ok {
			return c
		}
	}
	return nil
}

// ClassFQN returns the FQN of a class declaration ("" for anonymous classes).
func (e *Env) ClassFQN(c *syntax.ClassLike) string {
	if c == nil || c.Name == nil {
		return ""
	}
	if ns := e.Names.Namespace(c.Span().Start); ns != "" {
		return ns + `\` + c.Name.Value
	}
	return c.Name.Value
}

// classRef resolves a class reference expression (Name or expression) to a FQN.
func (e *Env) classRef(x syntax.Expr) string {
	switch n := x.(type) {
	case *syntax.Name:
		low := strings.ToLower(n.Value)
		switch low {
		case "self", "static":
			return e.ClassFQN(EnclosingClass(n))
		case "parent":
			if c := EnclosingClass(n); c != nil && len(c.Extends) > 0 && c.ClassKind != syntax.KindInterface {
				return e.Names.Class(c.Extends[0].Value, c.Span().Start)
			}
			return ""
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

func (e *Env) newType(n *syntax.New) types.Type {
	if _, ok := n.Class.(*syntax.ClassLike); ok {
		return types.Of("object")
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
	dt := types.FromDoc(doc, nil)
	if d.IsUnknown() || (d.Has("array") && !dt.IsUnknown()) || d.Has("mixed") || strictSuperset(dt, d) {
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
	for _, cls := range recv.Classes() {
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
	if m.Return == "" && m.DocReturn == "" {
		return e.methodBodyReturn(m, virtual)
	}
	return bindStatic(memberType(m.Return, m.DocReturn), cls)
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
	for _, cls := range recv.Classes() {
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
		if c := e.ClassFQN(EnclosingClass(n)); c != "" {
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
	if t, ok := e.overrideType(n); ok {
		return t
	}
	f := e.ResolveFunction(n)
	if f == nil {
		return types.Unknown
	}
	if f.Tpl != nil {
		if t, ok := e.tplReturn(f.Tpl, f.Params, n.Args, f.Return, nil, nil); ok {
			return t
		}
	}
	if f.Return == "" && f.DocReturn == "" {
		return e.BodyReturnType(f)
	}
	return memberType(f.Return, f.DocReturn)
}

// overrideType refines builtin return types that depend on arguments.
func (e *Env) overrideType(n *syntax.FuncCall) (types.Type, bool) {
	name, ok := n.Name.(*syntax.Name)
	if !ok || n.Args == nil {
		return types.Unknown, false
	}
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
	case "str_replace", "str_ireplace", "preg_replace", "preg_replace_callback", "substr_replace":
		subjectIdx := 2
		if strings.HasPrefix(strings.ToLower(fqn), "preg_replace_callback") || strings.ToLower(fqn) == "substr_replace" {
			subjectIdx = 0
			if strings.HasPrefix(strings.ToLower(fqn), "preg_replace_callback") {
				subjectIdx = 2
			}
		}
		if s := arg(subjectIdx); s != nil {
			st := e.TypeOf(s)
			if st.OnlyOf("string", "int", "float", "bool") { // scalar subjects are converted to string
				if strings.HasPrefix(strings.ToLower(fqn), "preg_") {
					return types.Of("string", "null"), true
				}
				return types.String, true
			}
			if st.IsArrayLike() {
				return types.Array, true
			}
		}
	case "current", "reset", "end", "next", "prev", "array_pop", "array_shift":
		if a := arg(0); a != nil {
			at := e.TypeOf(a)
			el := at.Elem()
			if el.IsUnknown() {
				el = e.shapeElem(at, a)
			} else if v, ok := unparen(a).(*syntax.Variable); ok && v.Name != "" && v.Name != "this" {
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
						if v, ok := unparen(a).(*syntax.Variable); ok && v.Name != "" && !e.pointerMoved(scopeOf(v), v.Name) {
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
		if c, ok := unparen(arg(0)).(*syntax.ConstFetch); ok && c.Name != nil {
			switch strings.ToLower(strings.TrimPrefix(c.Name.Value, `\`)) {
			case "false":
				return types.Of("int[]", "false"), true
			case "true":
				return types.Of("int", "float", "false"), true
			}
		}
	case "parse_url":
		// Without a component the result is the parts array (or false).
		switch len(n.Args.Args) {
		case 1:
			if arg(0) != nil {
				return types.Of("array", "false"), true
			}
		case 2:
			if c, ok := unparen(arg(1)).(*syntax.ConstFetch); ok && c != nil {
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
				case "array_slice", "array_filter":
					return t.WithoutShape().WithNonEmpty(false), true
				}
				return t.WithoutShape(), true
			}
		}
	}
	return types.Unknown, false
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
	return e.widenVarElem(el, v)
}

// widenVarElem unions el with every value written into the elements of
// variable v by a write reaching v (see widenElem). A nested write into an
// element makes the result unknown.
func (e *Env) widenVarElem(el types.Type, v *syntax.Variable) types.Type {
	ws, back := e.reachingWrites(v)
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

// scopeOf returns the function-like node (or nil for file scope) owning n.
func scopeOf(n syntax.Node) syntax.Node {
	for p := n.Parent(); p != nil; p = p.Parent() {
		switch p.(type) {
		case *syntax.Function, *syntax.Method, *syntax.Closure, *syntax.ArrowFunction:
			return p
		}
	}
	return nil
}

func (e *Env) variableType(v *syntax.Variable) types.Type {
	if v.Name == "" {
		return types.Unknown
	}
	if v.Name == "this" {
		if fqn := e.ClassFQN(EnclosingClass(v)); fqn != "" {
			return types.Of(`\` + fqn)
		}
		return types.Unknown
	}
	scope := scopeOf(v)
	sv := e.scopeVars(scope)
	defs := sv.defs[v.Name]
	if af, ok := scope.(*syntax.ArrowFunction); ok && len(defs) == 0 {
		// Arrow functions capture the enclosing scope by value.
		_ = af
		outer := e.scopeVars(scopeOf(scope))
		defs = outer.defs[v.Name]
	}
	if len(defs) > maxVarDefs {
		return types.Unknown
	}
	fwd, back, from := e.reaching(defs, v, scope)
	ts := make([]types.Type, 0, len(fwd)+len(back))
	for _, d := range fwd {
		ts = append(ts, d.typ())
	}
	for _, d := range back {
		// An unknown back-edge type (often a cycle through this very use)
		// adds nothing: keep what the forward definitions say.
		if t := d.typ(); !t.IsUnknown() {
			ts = append(ts, t)
		}
	}
	if len(ts) == 0 {
		return types.Unknown
	}
	t := types.Union(ts...)
	if t.HasShape() || t.IsNonEmptyArray() {
		ms := scope
		if _, ok := scope.(*syntax.ArrowFunction); ok && len(sv.defs[v.Name]) == 0 {
			ms = scopeOf(scope)
		}
		if t.HasShape() && e.shapeClobbered(ms, v.Name) {
			t = t.WithoutShape()
		}
		if t.IsNonEmptyArray() && e.nonEmptyBroken(ms, v.Name, from, v) {
			t = t.WithNonEmpty(false)
		}
	}
	return e.narrow(t, v, scope)
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
	case *syntax.Method:
		params = s.Params
		if s.Body != nil {
			body = []syntax.Node{s.Body}
		}
	case *syntax.Closure:
		params, body = s.Params, []syntax.Node{s.Body}
		for _, u := range s.Uses {
			u := u
			add(u.Var.Name, u.Span().Start, func() types.Type {
				outer := e.scopeVars(scopeOf(scope))
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
		p := p
		add(p.Var.Name, p.Span().Start, func() types.Type { return e.paramType(scope, p) })
	}
	for _, b := range body {
		if b == nil {
			continue
		}
		syntax.Inspect(b, func(n syntax.Node) bool {
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
			case *syntax.Function, *syntax.Method, *syntax.ArrowFunction, *syntax.ClassLike:
				return n == scope // do not descend into nested scopes
			case *syntax.Assign:
				e.collectDimWrites(n, n, n.Var, sv)
				end := n.Span().End
				var kill syntax.Span
				if _, plain := n.Var.(*syntax.Variable); plain && n.Op.Kind == syntax.TEqual {
					if es, ok := n.Parent().(*syntax.ExprStmt); ok {
						if blk, ok := es.Parent().(*syntax.Block); ok {
							kill = blk.Span()
						}
					}
				}
				e.collectTargets(n.Var, n.Span().Start, func() types.Type { return e.TypeOf(n) }, func(name string, pos uint32, t func() types.Type) {
					if name != "" {
						sv.defs[name] = append(sv.defs[name], varDef{pos: pos, end: end, kill: kill, typ: t})
					}
				})
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
	e.inlineVarDocs(scope, func(name string, pos, end uint32, t func() types.Type) {
		sv.defs[name] = append(sv.defs[name], varDef{pos: pos, docEnd: end, typ: t, doc: true})
	})
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

// collectItemTargets registers destructuring items; an item reads the
// destructured value's shape key (explicit literal key or position).
func (e *Env) collectItemTargets(items []*syntax.ArrayItem, pos uint32, t func() types.Type, add func(string, uint32, func() types.Type)) {
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
			item = shapeTarget(t, key)
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
	if d := e.DocOf(scope); d != nil {
		for _, dp := range d.Params() {
			if dp.Name == p.Var.Name {
				doc = dp.Type
			}
		}
	}
	if doc != "" {
		dt := types.FromDoc(doc, e.resolverFor(scope, at))
		if declared.IsUnknown() || declared.Has("array") || declared.Has("mixed") {
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
	for _, t := range toks[first:] {
		if t.Start >= span.End {
			break
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
