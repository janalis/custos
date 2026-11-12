package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

// PropertyMismatch is an assignment whose value every declaration of the
// written property rejects.
type PropertyMismatch struct {
	// Value is the assigned operand (the right-hand side).
	Value syntax.Expr
	// Property names the nominal declaration (`Class::$name`).
	Property string
	// Expected is the rejecting type of the nominal declaration, as declared.
	Expected string
	// Given is the type of the value the assignment stores.
	Given types.Type
	// Proven tells that the native property types reject the value's proven
	// type (see infer.Env.ProvenTypeOf); otherwise a PHPDoc type rejects it,
	// or the native types reject only its PHPDoc-aware type.
	Proven bool
}

// Message is the finding's text: the type to assign instead.
func (m PropertyMismatch) Message() string {
	return "Assign " + declaredTypeName(m.Expected) + " to " + m.Property + ", not " + typeName(m.Given) + "."
}

type propertyTypesKey struct{ assign syntax.Node }

// PropertyAssignmentMismatch reports the assignment n (plain, compound or
// `??=`) to a declared property when every declaration the write may reach
// (see propertyTargets) rejects the stored value (see types.AllRejected).
// Cached per assignment for the file.
func PropertyAssignmentMismatch(ctx *analysis.Context, n syntax.Node) (PropertyMismatch, bool) {
	m := ctx.Shared(propertyTypesKey{n}, func() any {
		m, ok := propertyAssignmentMismatch(ctx, n)
		if !ok {
			return (*PropertyMismatch)(nil)
		}
		return &m
	}).(*PropertyMismatch)
	if m == nil {
		return PropertyMismatch{}, false
	}
	return *m, true
}

func propertyAssignmentMismatch(ctx *analysis.Context, n syntax.Node) (PropertyMismatch, bool) {
	a, ok := n.(*syntax.Assign)
	if !ok || a.ByRef || a.Value == nil || !propertyTarget(a.Var) || reboundThis(a) {
		return PropertyMismatch{}, false
	}
	// The stored value: the result of a compound operation, the right-hand
	// side of a plain assignment or of `??=` (stored only when it runs).
	stored := syntax.Expr(a)
	if a.Op.Kind == syntax.TEqual || a.Op.Kind == syntax.TCoalesceEqual {
		stored = a.Value
	}
	given := ctx.TypeOf(stored)
	if given.IsUnknown() {
		return PropertyMismatch{}, false
	}
	label, targets, ok := writtenProperties(ctx, a)
	if !ok {
		return PropertyMismatch{}, false
	}
	check := fileTypeCheck(ctx)
	expected, allNative := "", true
	for i, p := range targets {
		native := check.declaredType(p.Type, "")
		var shown string
		switch {
		case types.AllRejected(native, given, check.weak):
			shown = p.Type
		case check.docTypeRejects(p.DocType, p.Default, native, given, check.weak):
			shown, allNative = p.DocType, false
		default:
			return PropertyMismatch{}, false
		}
		if i == 0 {
			expected = shown
		}
	}
	m := PropertyMismatch{Value: a.Value, Property: label, Expected: expected, Given: given}
	if allNative {
		proven := ctx.Types().ProvenTypeOf(stored)
		m.Proven = true
		for _, p := range targets {
			if !types.AllRejected(check.declaredType(p.Type, ""), proven, check.weak) {
				m.Proven = false
				break
			}
		}
		if m.Proven {
			m.Given = proven
		}
	}
	return m, true
}

// writtenProperties lists the declarations of the property assignment a
// writes (see propertyTarget), and names the nominal one: an instance property through a
// receiver whose native type names only classes, or a static property
// through a named class, `self`, `parent` or `static`.
func writtenProperties(ctx *analysis.Context, a *syntax.Assign) (string, []*index.Property, bool) {
	w := propertyWrite{ctx: ctx, assign: a}
	if t, ok := syntax.UnwrapParens(a.Var).(*syntax.PropertyFetch); ok {
		// Only native declarations decide which class is written.
		typ := ctx.Types().ProvenTypeOf(t.Var).Without("null")
		classes := typ.Classes()
		if len(classes) == 0 || len(classes) != len(typ.Atoms()) {
			return "", nil, false
		}
		w.name, w.this = t.Name.(*syntax.Identifier).Value, isThis(t.Var)
		label := ""
		var all []*index.Property
		for _, cls := range classes {
			l, ps, ok := w.targets(strings.TrimPrefix(cls, `\`), true)
			if !ok {
				return "", nil, false
			}
			if label == "" {
				label = l
			}
			all = append(all, ps...)
		}
		return label, all, true
	}
	t := syntax.UnwrapParens(a.Var).(*syntax.StaticPropertyFetch)
	name, named := t.Class.(*syntax.Name)
	if !named {
		return "", nil, false
	}
	cls := StaticCallClass(ctx, name)
	if cls == "" {
		return "", nil, false
	}
	w.name, w.static = t.Name.(*syntax.Variable).Name, true
	return w.targets(cls, strings.EqualFold(name.Value, "static"))
}

// propertyTarget reports whether x writes a property named in the source.
func propertyTarget(x syntax.Expr) bool {
	switch t := syntax.UnwrapParens(x).(type) {
	case *syntax.PropertyFetch:
		_, named := t.Name.(*syntax.Identifier)
		return named && !t.NullSafe
	case *syntax.StaticPropertyFetch:
		v, named := t.Name.(*syntax.Variable)
		return named && v.Name != ""
	}
	return false
}

func isThis(x syntax.Expr) bool {
	v, ok := syntax.UnwrapParens(x).(*syntax.Variable)
	return ok && v.Name == "this"
}

// propertyWrite is a write to the property name: static or not, through
// `$this` or not.
type propertyWrite struct {
	ctx          *analysis.Context
	assign       *syntax.Assign
	name         string
	static, this bool
}

// targets lists the declaration of the property as class cls resolves it
// and, with subtypes, as every indexed subtype of cls resolves it (unless
// the property is private or final, or the class final); every subtype
// must also be free of `__set`.
func (w propertyWrite) targets(cls string, subtypes bool) (string, []*index.Property, bool) {
	ix := w.ctx.Index()
	c := ix.Class(cls, w.ctx.PHP)
	if c == nil || c.Kind == syntax.KindTrait {
		return "", nil, false
	}
	p, ok := w.declaration(cls)
	if !ok || !w.writable(p) {
		return "", nil, false
	}
	label := strings.TrimPrefix(c.FQN, `\`)
	if owner := propertyOwner(p); owner != "" {
		label = strings.TrimPrefix(owner, `\`)
	}
	label += "::$" + p.Name
	out := []*index.Property{p}
	if !subtypes || c.Final {
		return label, out, true
	}
	d := certainDescendants(ix, cls, w.ctx.PHP)
	if !d.certain {
		return "", nil, false
	}
	redeclarable := p.Visibility != index.Private && !p.Final
	for _, child := range d.classes {
		sub, ok := w.declaration(child)
		if !ok {
			return "", nil, false
		}
		if redeclarable {
			out = append(out, sub)
		}
	}
	return label, out, true
}

// declaration resolves the property on class cls when that declaration is
// certain and stores what is assigned: the class and the declaring one are
// declared once, every ancestor is indexed, PHP itself does not declare it,
// no hook, `@property` tag or `__set` handler intervenes.
func (w propertyWrite) declaration(cls string) (*index.Property, bool) {
	ix := w.ctx.Index()
	ver := w.ctx.PHP
	if len(ix.ClassDecls(cls, ver)) != 1 || !HierarchyResolved(ix, cls, ver) {
		return nil, false
	}
	p := ix.FindProperty(cls, w.name, ver)
	if p == nil || p.Magic || p.Builtin || p.Static != w.static || p.Hooked || p.WritesRunCode || len(ix.ClassDecls(p.Class, ver)) != 1 {
		return nil, false
	}
	if !w.static && ix.FindMethod(cls, "__set", ver) != nil {
		return nil, false
	}
	return p, true
}

// propertyOwner is the class whose objects hold p: the class using the
// trait that declares it, or the declaring class.
func propertyOwner(p *index.Property) string {
	if p.TypeClass != "" {
		return p.TypeClass
	}
	return p.Class
}

// writable reports whether the write reaches p's storage rather than
// failing before the type check: p is visible for writing from the
// enclosing class (code in a closure, which may be bound to another scope,
// writes public properties only), and a readonly property is initialized
// once, by the only write in its own class's constructor.
func (w propertyWrite) writable(p *index.Property) bool {
	owner := strings.TrimPrefix(propertyOwner(p), `\`)
	scope := ""
	if c := syntax.EnclosingClass(w.assign); c != nil && c.ClassKind != syntax.KindTrait {
		scope = strings.TrimPrefix(w.ctx.Types().ClassFQN(c), `\`)
	}
	switch fn := syntax.EnclosingFuncLike(w.assign).(type) {
	case *syntax.Closure, *syntax.ArrowFunction:
		scope = ""
	case *syntax.Method:
		if p.Readonly {
			return !p.Promoted && w.this && strings.EqualFold(fn.Name.Value, "__construct") &&
				strings.EqualFold(scope, owner) && onlyWrite(fn, p.Name)
		}
	}
	if p.Readonly {
		return false
	}
	switch p.WriteVisibility(w.ctx.PHP) {
	case index.Private:
		return scope != "" && strings.EqualFold(scope, owner)
	case index.Protected:
		ix := w.ctx.Index()
		return scope != "" && (ix.IsSubtype(scope, owner, w.ctx.PHP) || ix.IsSubtype(owner, scope, w.ctx.PHP))
	}
	return true
}

// onlyWrite reports whether method fn assigns `$this->name` once.
func onlyWrite(fn *syntax.Method, name string) bool {
	writes := 0
	syntax.Inspect(fn.Body, func(n syntax.Node) bool {
		if a, ok := n.(*syntax.Assign); ok {
			if f, ok := syntax.UnwrapParens(a.Var).(*syntax.PropertyFetch); ok && isThis(f.Var) {
				if id, ok := f.Name.(*syntax.Identifier); ok && id.Value == name {
					writes++
				}
			}
		}
		return true
	})
	return writes == 1
}
