package infer

import (
	"slices"
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

// ClassFQN returns the semantic identity of a class declaration.
func (e *Env) ClassFQN(c *syntax.ClassLike) string { return e.Names.SymbolFQN(c) }

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

// traitThis returns the FQN of the trait x is `$this` in, "" when x is not
// `$this` inside a trait.
func (e *Env) traitThis(x syntax.Expr) string {
	v, ok := syntax.UnwrapParens(x).(*syntax.Variable)
	c := syntax.EnclosingClass(x)
	if !ok || v.Name != "this" || c == nil || c.ClassKind != syntax.KindTrait {
		return ""
	}
	return e.ClassFQN(c)
}

// traitRequired types `$this` inside a trait as the classes and interfaces
// the trait requires of the classes using it (@require-extends,
// @require-implements): every such class is one of them. Unknown without
// any.
func (e *Env) traitRequired(this *syntax.Variable) types.Type {
	fqn := e.traitThis(this)
	if fqn == "" {
		return types.Unknown
	}
	c := e.Index.Class(fqn, e.PHP)
	if c == nil || len(c.Requires) == 0 {
		return types.Unknown
	}
	names := make([]string, len(c.Requires))
	for i, r := range c.Requires {
		names[i] = `\` + r
	}
	return types.FromDoc(strings.Join(names, "&"), nil)
}

// traitPropertyType types `$this->p` inside a trait, where `$this` has no
// class of its own, from the trait's declaration of p: a class using the
// trait cannot declare p again with another type (unlike its methods, which
// the class may replace). Unknown otherwise, and when the type names the
// trait itself (`self` there stands for the using class).
func (e *Env) traitPropertyType(n *syntax.PropertyFetch) types.Type {
	fqn := e.traitThis(n.Var)
	id, named := n.Name.(*syntax.Identifier)
	if fqn == "" || !named {
		return types.Unknown
	}
	p := e.Index.FindProperty(fqn, id.Value, e.PHP)
	if p == nil || p.Magic || p.Static {
		return types.Unknown
	}
	t := e.propType(p, fqn)
	if t.Has(`\` + fqn) {
		return types.Unknown
	}
	return t
}

// traitMethodType types `$this->m()` inside a trait from the return type
// the trait declares for m. A class using the trait may replace m with
// another signature, so the native environment, used for proofs, keeps
// the call unknown; so does a type naming the trait itself.
func (e *Env) traitMethodType(n *syntax.MethodCall, name string) types.Type {
	fqn := e.traitThis(n.Var)
	if fqn == "" || e.native {
		return types.Unknown
	}
	if m := e.Index.FindMethod(fqn, name, e.PHP); m == nil || m.Magic {
		return types.Unknown
	}
	t := e.methodReturn(fqn, name, true, fqn, nil, n.Args)
	if t.Has(`\` + fqn) {
		return types.Unknown
	}
	return t
}

// classRef resolves a class reference expression (Name or expression) to a FQN.
func (e *Env) classRef(x syntax.Expr) string {
	switch n := x.(type) {
	case *syntax.ClassLike:
		return e.ClassFQN(n)
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

// anonClassType retains the declaration identity so its own members and
// inherited contracts use the same indexed lookup as named classes.
func (e *Env) anonClassType(cl *syntax.ClassLike) types.Type {
	return types.Of(`\` + e.ClassFQN(cl))
}

func (e *Env) newType(n *syntax.New) types.Type {
	if cl, ok := n.Class.(*syntax.ClassLike); ok {
		return e.anonClassType(cl)
	}
	if cls := e.classRef(n.Class); cls != "" {
		return e.ctorGenerics(types.Of(`\`+cls), cls, n.Args)
	}
	if cls := e.classStringRef(n.Class); cls != "" {
		return types.Of(`\` + cls) // the class or one extending it
	}
	return types.Unknown
}

// classStringRef returns class C when expression x (not a name) holds a
// `class-string<C>` of a single class: it names C or a class extending it.
func (e *Env) classStringRef(x syntax.Expr) string {
	if _, named := x.(*syntax.Name); named {
		return ""
	}
	if cs := types.ClassStringOf(e.TypeOf(x)).Classes(); len(cs) == 1 {
		return strings.TrimPrefix(cs[0], `\`)
	}
	return ""
}

// ctorGenerics gives t, the type `new cls(…)` creates, the class template
// arguments its own constructor binds from args (`new Box($foo)` with
// `@param T $value` is `Box<Foo>`); templates left unbound take their bound,
// else mixed. t is returned unchanged when nothing binds.
func (e *Env) ctorGenerics(t types.Type, cls string, args *syntax.ArgList) types.Type {
	c := e.Index.Class(cls, e.PHP)
	if c == nil || len(c.Templates) == 0 {
		return t
	}
	m := c.Methods["__construct"]
	if m == nil || m.CtorTpl == nil {
		return t
	}
	mb := e.bindCall(&index.FuncTemplates{Templates: c.Templates, Params: m.CtorTpl}, m.Params, args)
	if len(mb) == 0 {
		return t
	}
	targs := make([]types.Type, len(c.Templates))
	for i, tp := range c.Templates {
		switch bt, ok := mb[tp.Name]; {
		case ok:
			targs[i] = bt
		case tp.Bound != "":
			targs[i] = types.FromDoc(tp.Bound, nil)
		}
	}
	return t.WithTypeArgs(`\`+cls, targs)
}

// bindStatic binds types whose declaration and receiver share a class context.
func bindStatic(t types.Type, receiver string) types.Type {
	return t.BindRelative(receiver, "", receiver)
}

// bindMember distinguishes a declaration's self/parent from late static binding.
func (e *Env) bindMember(t types.Type, declaration, effective, receiver string) types.Type {
	if !t.HasRelative() {
		return t
	}
	owner := declaration
	if effective != "" {
		owner = effective
	}
	parent := ""
	if c := e.Index.Class(owner, e.PHP); c != nil {
		if c.Kind == syntax.KindTrait {
			owner = ""
		} else {
			parent = c.Parent
		}
	} else {
		owner = ""
	}
	return t.BindRelative(owner, parent, receiver)
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
		return e.bindMember(types.FromDoc(p.Type, nil), p.Class, p.TypeClass, receiver)
	}
	if p.Type == "" && p.DocType == "" {
		return e.bindMember(e.inferredProp(p), p.Class, p.TypeClass, receiver)
	}
	return e.bindMember(e.memberType(p.Type, p.DocType), p.Class, p.TypeClass, receiver)
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
	bind := func(t types.Type) types.Type {
		return e.bindMember(t, m.Class, m.TypeClass, origin)
	}
	if e.userDoc(m.Builtin) {
		return bind(types.FromDoc(m.Return, nil))
	}
	if m.Tpl != nil {
		var classB tplBindings
		c := e.Index.Class(m.Class, e.PHP)
		if c != nil && len(c.Templates) > 0 {
			classB = e.genBindings(origin, args)[strings.ToLower(strings.TrimPrefix(m.Class, `\`))]
		}
		if t, ok := e.tplReturn(m.Tpl, m.Params, call, m.Return, classB, c); ok {
			return bind(t)
		}
	}
	if t, ok := e.genMethodReturn(m, origin, args); ok {
		return bind(t)
	}
	if m.CondReturn != "" {
		if t, ok := e.condCall(m.CondReturn, m.Class, m.Tpl, m.Params, call, m.Return, m.DocReturn); ok {
			return bind(t)
		}
	}
	if m.Return == "" && m.DocReturn == "" {
		// An override without return type keeps the contract of the method
		// it overrides (a parent's or interface's `: mixed`); the body is
		// only consulted when no ancestor declares one.
		if pm := e.inheritedSignature(m); pm != nil {
			if pm.Builtin {
				return e.bindMember(e.builtinMemberType(pm.Return, pm.DocReturn), pm.Class, pm.TypeClass, origin)
			}
			return e.bindMember(e.memberType(pm.Return, pm.DocReturn), pm.Class, pm.TypeClass, origin)
		}
		return e.methodBodyReturn(m, virtual)
	}
	if m.Builtin {
		return bind(e.builtinMemberType(m.Return, m.DocReturn))
	}
	if m.DocReturn == "" && types.FromDoc(m.Return, nil).HasAny("array", "iterable") {
		// An override declaring only `: array` keeps the documented return
		// of the method it overrides (`{@inheritDoc}`, or no doc at all), as
		// a doc narrows its own declaration (see pickMemberType).
		if pm := e.inheritedDoc(m); pm != nil {
			if t, ok := e.genMethodReturn(pm, origin, args); ok {
				return bind(t)
			}
			return e.bindMember(e.memberType(m.Return, pm.DocReturn), pm.Class, pm.TypeClass, origin)
		}
	}
	if t := e.methodBodyReturn(m, virtual); !t.IsUnknown() {
		return bind(t) // a plain declared array, refined by the body
	}
	return bind(e.memberType(m.Return, m.DocReturn))
}

// inheritedDoc returns the nearest method m overrides (in the parents and
// interfaces of its class) that documents a return type; nil when none
// does.
func (e *Env) inheritedDoc(m *index.Method) *index.Method {
	anc := e.Index.Ancestors(strings.TrimPrefix(m.Class, `\`), e.PHP)
	low := strings.ToLower(m.Name)
	for _, c := range anc[min(1, len(anc)):] {
		if pm, ok := c.Methods[low]; ok && !pm.Magic && pm.DocReturn != "" {
			return pm
		}
	}
	return nil
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

// inheritedParamDoc returns the documented type of parameter i of method m
// taken from the nearest overridden method (in the parents and interfaces of
// its class) that documents the parameter at the same position; parameters
// may be renamed, so the name is never matched. Unknown when none does, and
// for variadic parameters.
func (e *Env) inheritedParamDoc(m *syntax.Method, i int) types.Type {
	if i >= len(m.Params) || m.Params[i].Variadic {
		return types.Unknown
	}
	c := syntax.EnclosingClass(m)
	if c == nil || c.ClassKind == syntax.KindTrait {
		return types.Unknown
	}
	child := e.ClassFQN(c)
	anc := e.Index.Ancestors(child, e.PHP)
	low := strings.ToLower(m.Name.Value)
	for _, a := range anc[min(1, len(anc)):] {
		pm, ok := a.Methods[low]
		if !ok || pm.Magic || pm.Visibility == index.Private || i >= len(pm.Params) || pm.Params[i].Variadic || pm.Params[i].DocType == "" {
			continue
		}
		return e.bindMember(types.FromDoc(e.resolvedDoc(pm.Params[i].DocType), nil), pm.Class, pm.TypeClass, `\`+child)
	}
	return types.Unknown
}

func (e *Env) methodCallType(n *syntax.MethodCall) types.Type {
	id, ok := n.Name.(*syntax.Identifier)
	if !ok {
		return types.Unknown
	}
	// A trait's own method first: it replaces the one of a class it
	// requires the using class to extend.
	if t := e.traitMethodType(n, id.Value); !t.IsUnknown() {
		return t
	}
	recv := e.chainReceiver(n.Var, n.NullSafe)
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
	return t
}

func (e *Env) staticCallType(n *syntax.StaticCall) types.Type {
	id, ok := n.Name.(*syntax.Identifier)
	if !ok {
		return types.Unknown
	}
	cls := e.classRef(n.Class)
	if cls == "" {
		cls = e.classStringRef(n.Class)
	}
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

// constFixed reports whether a fetch of constant k through class reference
// ref (resolved to cls) always reads k's own value: a subclass may override
// a constant that `static::` or an object reaches, unless the constant or
// the class is final.
func (e *Env) constFixed(ref syntax.Expr, cls string, k *index.ClassConst) bool {
	if nm, ok := ref.(*syntax.Name); (ok && !strings.EqualFold(nm.Value, "static")) || k.Final {
		return true
	}
	c := e.Index.Class(cls, e.PHP)
	return c != nil && (c.Final || c.Kind == syntax.KindEnum)
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
	if c, name, ok := e.enumCase(n); ok {
		return types.EnumCases(`\`+c.FQN, name)
	}
	cls := e.classRef(n.Class)
	if cls == "" {
		return types.Unknown
	}
	k := e.Index.FindConst(cls, id.Value, e.PHP)
	if k == nil {
		return types.Unknown
	}
	if k.ValueType != "" && e.constFixed(n.Class, cls, k) {
		return types.FromDoc(k.ValueType, nil)
	}
	if k.Type != "" {
		t := types.FromDoc(k.Type, nil)
		owner := k.Class
		if k.TypeClass != "" {
			owner = k.TypeClass
		}
		if t.HasAny("self", "parent") {
			c := e.Index.Class(owner, e.PHP)
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
		return bindStatic(t, owner)
	}
	return literalTextType(k.Value)
}

// resolvedDoc retains cheap normalized member typing for ordinary contracts;
// constant expressions are deferred until the complete project index exists.
func (e *Env) resolvedDoc(doc string) string {
	if !strings.Contains(doc, "::") && !strings.Contains(doc, "key-of<") && !strings.Contains(doc, "value-of<") {
		return doc
	}
	return e.docType(doc, nil).DocString()
}

func (e *Env) memberType(declared, doc string) types.Type {
	doc = e.resolvedDoc(doc)
	if narrowed, ok := e.narrowerDoc(types.FromDoc(declared, nil), types.FromDoc(doc, nil)); ok {
		return narrowed
	}
	return memberType(declared, doc)
}

// narrowerDoc returns the doc type dt of a member declared d when dt only
// names classes that extend or implement every class d names (`@var
// Message` on a `MessageInterface` property), with null when d admits it:
// the doc promises a subtype of what the declaration allows.
func (e *Env) narrowerDoc(d, dt types.Type) (types.Type, bool) {
	declared, documented := d.Without("null"), dt.Without("null")
	if !classesOnly(declared) || !classesOnly(documented) || dt.Has("null") && !d.Has("null") {
		return types.Type{}, false
	}
	for _, sub := range documented.Atoms() {
		if declared.Has(sub) {
			return types.Type{}, false
		}
		ancestors := e.Index.Ancestors(strings.TrimPrefix(sub, `\`), e.PHP)
		for _, super := range declared.Atoms() {
			if !slices.ContainsFunc(ancestors, func(c *index.Class) bool {
				return strings.EqualFold(`\`+strings.TrimPrefix(c.FQN, `\`), super)
			}) {
				return types.Type{}, false
			}
		}
	}
	if d.Has("null") {
		return types.Union(documented, types.Of("null")), true
	}
	return documented, true
}

// classesOnly reports a type naming one class or more and nothing else.
func classesOnly(t types.Type) bool {
	for _, a := range t.Atoms() {
		if !strings.HasPrefix(a, `\`) || strings.HasSuffix(a, "[]") {
			return false
		}
	}
	return !t.IsUnknown() && len(t.Intersection()) == 0
}

func (e *Env) builtinMemberType(declared, doc string) types.Type {
	return builtinMemberType(declared, e.resolvedDoc(doc))
}
