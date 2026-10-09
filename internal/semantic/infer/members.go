package infer

import (
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
		return types.Of(`\` + cls)
	}
	return types.Unknown
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
	return e.bindMember(memberType(p.Type, p.DocType), p.Class, p.TypeClass, receiver)
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
		if t, ok := e.condCall(m.CondReturn, m.Params, call, m.Return, m.DocReturn); ok {
			return bind(t)
		}
	}
	if m.Return == "" && m.DocReturn == "" {
		// An override without return type keeps the contract of the method
		// it overrides (a parent's or interface's `: mixed`); the body is
		// only consulted when no ancestor declares one.
		if pm := e.inheritedSignature(m); pm != nil {
			if pm.Builtin {
				return e.bindMember(builtinMemberType(pm.Return, pm.DocReturn), pm.Class, pm.TypeClass, origin)
			}
			return e.bindMember(memberType(pm.Return, pm.DocReturn), pm.Class, pm.TypeClass, origin)
		}
		return e.methodBodyReturn(m, virtual)
	}
	if m.Builtin {
		return bind(builtinMemberType(m.Return, m.DocReturn))
	}
	return bind(memberType(m.Return, m.DocReturn))
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
