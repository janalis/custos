package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/types"
)

// PropertyMismatch is an assignment whose value every declaration of the
// written property rejects.
type PropertyMismatch struct {
	// Value is the range reported: the assigned operand (the right-hand
	// side), the target a destructuring or foreach loop writes, or ++/--.
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

type propertyTypesKey struct{ write syntax.Node }

// PropertyWriteMismatches reports the writes node n makes to declared
// properties when every declaration a write may reach (see
// propertyTargets) rejects the stored value (see types.AllRejected): an
// assignment (plain, compound, `??=` or by reference), a destructuring
// assignment, a foreach loop binding its key or value to properties, or
// ++/--, the only kinds n may have. Cached per node for the file.
func PropertyWriteMismatches(ctx *analysis.Context, n syntax.Node) []PropertyMismatch {
	return ctx.Shared(propertyTypesKey{n}, func() any {
		switch n := n.(type) {
		case *syntax.Assign:
			if m, ok := propertyAssignmentMismatch(ctx, n); ok {
				return []PropertyMismatch{m}
			}
			return destructuringMismatches(ctx, n)
		case *syntax.Foreach:
			return foreachMismatches(ctx, n)
		}
		return incDecMismatches(ctx, n.(*syntax.IncDec))
	}).([]PropertyMismatch)
}

func propertyAssignmentMismatch(ctx *analysis.Context, n syntax.Node) (PropertyMismatch, bool) {
	a, ok := n.(*syntax.Assign)
	if !ok || a.Value == nil || !propertyTarget(a.Var) || reboundThis(a) {
		return PropertyMismatch{}, false
	}
	if a.ByRef {
		return referenceMismatch(ctx, a)
	}
	// The stored value: the result of a compound operation, the right-hand
	// side of a plain assignment or of `??=` (stored only when it runs).
	stored := syntax.Expr(a)
	plain := a.Op.Kind == syntax.TEqual || a.Op.Kind == syntax.TCoalesceEqual
	if plain {
		stored = a.Value
	}
	w := propertyStore{target: a.Var, at: a, report: a.Value, modifies: !plain}
	return judgeStore(ctx, w, func(env *infer.Env) types.Type { return env.TypeOf(stored) })
}

// propertyStore is one value stored in the property target names, by the
// writing node at; report is the range a finding covers; modifies tells a
// write that cannot initialize a readonly property: it reads the property
// first (compound assignment, ++/--) or binds a reference.
type propertyStore struct {
	target, report syntax.Expr
	at             syntax.Node
	modifies       bool
}

// judgeStore reports store when every declaration it may reach rejects the
// value typed by stored, run in the PHPDoc-aware environment and, for a
// finding resting on native types only, in the proven one.
func judgeStore(ctx *analysis.Context, w propertyStore, stored func(*infer.Env) types.Type) (PropertyMismatch, bool) {
	given := stored(ctx.Types())
	if given.IsUnknown() {
		return PropertyMismatch{}, false
	}
	label, targets, ok := writtenProperties(ctx, w)
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
		case check.docTypeRejects(documented(p.DocType, p.DocValues, given), propertyDefault(p), native, given, check.weak):
			shown, allNative = documented(p.DocType, p.DocValues, given), false
		default:
			return PropertyMismatch{}, false
		}
		if i == 0 {
			expected = shown
		}
	}
	m := PropertyMismatch{Value: w.report, Property: label, Expected: expected, Given: given}
	if allNative {
		proven := stored(ctx.Types().Proven())
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

// writtenProperties lists the declarations of the property store w writes
// (see propertyTarget), and names the nominal one: an instance property through a
// receiver whose native type names only classes, or a static property
// through a named class, `self`, `parent` or `static`.
func writtenProperties(ctx *analysis.Context, s propertyStore) (string, []*index.Property, bool) {
	w := propertyWrite{ctx: ctx, at: s.at, modifies: s.modifies}
	if t, ok := syntax.UnwrapParens(s.target).(*syntax.PropertyFetch); ok {
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
	t := syntax.UnwrapParens(s.target).(*syntax.StaticPropertyFetch)
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
	ctx                    *analysis.Context
	at                     syntax.Node
	name                   string
	static, this, modifies bool
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

// propertyDefault is the default value of p as declared, and `null` for
// an untyped property declared without one, which PHP starts as null.
func propertyDefault(p *index.Property) string {
	if p.Type == "" && p.Default == "" && !p.Promoted {
		return "null"
	}
	return p.Default
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
// once, by the only write in its own class's constructor, which does not
// read it first.
func (w propertyWrite) writable(p *index.Property) bool {
	owner := strings.TrimPrefix(propertyOwner(p), `\`)
	scope := ""
	if c := syntax.EnclosingClass(w.at); c != nil && c.ClassKind != syntax.KindTrait {
		scope = strings.TrimPrefix(w.ctx.Types().ClassFQN(c), `\`)
	}
	switch fn := syntax.EnclosingFuncLike(w.at).(type) {
	case *syntax.Closure, *syntax.ArrowFunction:
		scope = ""
	case *syntax.Method:
		if p.Readonly {
			return !p.Promoted && !w.modifies && w.this && strings.EqualFold(fn.Name.Value, "__construct") &&
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

// onlyWrite reports whether method fn writes `$this->name` once (see
// storeTargets).
func onlyWrite(fn *syntax.Method, name string) bool {
	writes := 0
	syntax.Inspect(fn.Body, func(n syntax.Node) bool {
		storeTargets(n, func(x syntax.Expr) {
			if f, ok := syntax.UnwrapParens(x).(*syntax.PropertyFetch); ok && isThis(f.Var) {
				if id, ok := f.Name.(*syntax.Identifier); ok && id.Value == name {
					writes++
				}
			}
		})
		return true
	})
	return writes == 1
}
