package util

import (
	"strings"

	"custos/internal/syntax"
)

// PossibleValuesComplete is a conservative variant of PossibleValues for
// consumers that must know *every* value an expression can take (for
// example to validate a format string). It returns the candidate value
// expressions of e and whether that set is complete; when complete is
// false the set must not be relied upon.
//
// Candidates are collected as in PossibleValues (parentheses stripped,
// both ternary branches — the condition for `?:` —, both `??` operands,
// anything else is a value itself), except that every source the file
// cannot fully account for makes the result incomplete:
//
//   - local variable: only the plain `=` assignments that may reach the use
//     (ReachingAssignments) count; incomplete in top-level code, when the
//     entry value may reach the use (parameter, closure import, undefined),
//     or when the variable has any other kind of write in its scope
//     (compound or by-reference assignment, destructuring, foreach, global,
//     static, catch, unset, ++/--) or is imported by reference into a
//     closure;
//   - `$this->p`, `self::$p`, `static::$p`, `C::$p` (C declared in this
//     file): only properties declared in the class itself (without hooks;
//     for `$this->p`, private or in a final or anonymous class, since a
//     subclass may redeclare it with another default);
//     the declared default (an untyped property without one is null, i.e.
//     incomplete) plus every plain `=` assignment to the property anywhere
//     in the class; any other write in the class makes the result
//     incomplete. Writes from outside the class (other files, subclasses)
//     are not tracked;
//   - class constant (`self::`, `C::` for a class in this file, `static::`
//     only in a final class) declared in the class itself: its value;
//   - global constant declared in this file (`const` / `define()`): its
//     value; true/false/null are values themselves;
//   - anything that cannot be resolved this way: incomplete.
func PossibleValuesComplete(f *syntax.File, e syntax.Expr) (vals []syntax.Expr, complete bool) {
	pc := &completeValues{valueWalk: newValueWalk(f)}
	pc.elvisCond, pc.nilStops = true, true
	pc.resolve = pc.resolveExpr
	pc.collect(e)
	return pc.result()
}

type completeValues struct{ valueWalk }

func (pc *completeValues) resolveExpr(e syntax.Expr) bool {
	switch x := e.(type) {
	case *syntax.Variable:
		pc.variable(x)
	case *syntax.PropertyFetch:
		id, ok := x.Name.(*syntax.Identifier)
		v, isVar := x.Var.(*syntax.Variable)
		if !ok || !isVar || v.Name != "this" || x.NullSafe {
			pc.stop = true
			return true
		}
		pc.property(syntax.EnclosingClass(x), id.Value, false)
	case *syntax.StaticPropertyFetch:
		v, ok := x.Name.(*syntax.Variable)
		if !ok || v.NameExpr != nil || v.Name == "" {
			pc.stop = true
			return true
		}
		pc.property(pc.staticClass(x, x.Class), v.Name, true)
	case *syntax.ClassConstFetch:
		pc.classConst(x)
	case *syntax.ConstFetch:
		pc.constant(x)
	default:
		return false
	}
	return true
}

func (pc *completeValues) variable(v *syntax.Variable) {
	if v.NameExpr != nil || v.Name == "" || v.Name == "this" {
		pc.stop = true
		return
	}
	scope := syntax.EnclosingFuncLike(v)
	if scope == nil {
		pc.stop = true
		return
	}
	_, body := ScopeParts(scope)
	if body == nil {
		pc.stop = true
		return
	}
	ix := assignsUnder(pc.f, body)
	if otherWrites(pc.f, scope)[v.Name] || ix.byRefUse[v.Name] || ix.unstable[v.Name] || len(ix.byVar[v.Name]) > maxPossibleValues {
		pc.stop = true
		return
	}
	defs, entry := ReachingAssignmentsIn(pc.f, scope, v, v.Name)
	if entry || len(defs) == 0 {
		pc.stop = true
		return
	}
	for _, d := range defs {
		pc.collect(AssignedValue(d))
	}
}

// staticClass resolves the class named by a static access (`self`,
// `static` — final classes only —, or a class declared in this file).
func (pc *completeValues) staticClass(at syntax.Node, class syntax.Expr) *syntax.ClassLike {
	cn, ok := class.(*syntax.Name)
	if !ok {
		return nil
	}
	switch strings.ToLower(cn.Value) {
	case "self":
		return syntax.EnclosingClass(at)
	case "static":
		if c := syntax.EnclosingClass(at); c != nil && c.Modifiers.Has(syntax.TFinal) {
			return c
		}
		return nil
	case "parent":
		return nil
	}
	return fileClassNamed(pc.f, LastNamePart(cn.Value))
}

func (pc *completeValues) property(class *syntax.ClassLike, name string, static bool) {
	if class == nil {
		pc.stop = true
		return
	}
	idx := classPropWrites(pc.f, class)
	d, ok := idx.decls[name]
	if !ok {
		pc.stop = true
		return
	}
	decl, prop := d.item, d.prop
	if prop.Modifiers.Has(syntax.TStatic) != static || len(prop.Hooks) > 0 || !static && mayRedeclare(class, prop) {
		pc.stop = true
		return
	}
	switch {
	case decl.Default != nil:
		pc.collect(decl.Default)
	case prop.Type == nil:
		pc.stop = true // implicit null
		return
	}
	w := idx.writes[propWriteKey{name, static}]
	if w != nil && w.dirty {
		pc.stop = true
		return
	}
	var vals []syntax.Expr
	if w != nil {
		vals = w.vals
	}
	if decl.Default == nil && len(vals) == 0 {
		pc.stop = true
		return
	}
	for _, v := range vals {
		pc.collect(v)
	}
}

// mayRedeclare reports whether a subclass may redeclare the instance
// property (with another default): `$this->p` then reads the subclass's
// value. Private properties and final or anonymous classes are safe.
func mayRedeclare(class *syntax.ClassLike, prop *syntax.Property) bool {
	return !prop.Modifiers.Has(syntax.TPrivate) && !class.Modifiers.Has(syntax.TFinal) && class.Name != nil
}

func (pc *completeValues) classConst(c *syntax.ClassConstFetch) {
	id, ok := c.Name.(*syntax.Identifier)
	class := pc.staticClass(c, c.Class)
	if !ok || class == nil {
		pc.stop = true
		return
	}
	if vals := classConstValues(class, id.Value); len(vals) > 0 {
		pc.collect(vals[0])
		return
	}
	pc.stop = true
}

func (pc *completeValues) constant(c *syntax.ConstFetch) {
	name := LastNamePart(c.Name.Value)
	switch strings.ToLower(name) {
	case "true", "false", "null":
		pc.out = append(pc.out, c)
		return
	}
	found := fileConstValue(pc.f, name)
	if found == nil {
		pc.stop = true
		return
	}
	pc.collect(found)
}

type otherWritesKey struct{ scope syntax.Node }

// otherWrites returns the names of scope having a write other than a plain
// `$v = …` assignment (computed once per scope and file; f is never nil
// here: PossibleValuesComplete needs the file).
func otherWrites(f *syntax.File, scope syntax.Node) map[string]bool {
	build := func() any {
		m := map[string]bool{}
		for name, accs := range VarAccessesByName(f, scope) {
			for _, acc := range accs {
				if !acc.Write {
					continue
				}
				a, ok := acc.By.(*syntax.Assign)
				if !ok || a.Op.Kind != syntax.TEqual || a.ByRef || syntax.UnwrapParens(a.Var) != syntax.Expr(acc.Var) {
					m[name] = true
					break
				}
			}
		}
		return m
	}
	return f.Memo(otherWritesKey{scope}, build).(map[string]bool)
}

type propWriteKey struct {
	name   string
	static bool
}

// propWrites records the writes to one property inside its class: the
// values of plain `=` assignments, in source order, and whether any other
// kind of write (compound, by-reference, ++/--) exists.
type propWrites struct {
	vals  []syntax.Expr
	dirty bool
}

type classPropWritesKey struct{ class *syntax.ClassLike }

type propDecl struct {
	item *syntax.PropertyItem
	prop *syntax.Property
}

// classProps is the per-class index behind property lookups: the last
// declaration of each property name and the writes to each property.
type classProps struct {
	decls  map[string]propDecl
	writes map[propWriteKey]*propWrites
}

// classPropWrites indexes, once per class and file, the writes to
// `$this->p` and to static properties of class (`self::$p`, `static::$p`
// in a final class, `C::$p`) found anywhere in its members. Building it per
// lookup made offset checks on huge classes quadratic.
func classPropWrites(f *syntax.File, class *syntax.ClassLike) *classProps {
	build := func() any {
		idx := map[propWriteKey]*propWrites{}
		decls := map[string]propDecl{}
		pc := &completeValues{valueWalk: newValueWalk(f)}
		keyOf := func(e syntax.Expr) (propWriteKey, bool) {
			switch x := syntax.UnwrapParens(e).(type) {
			case *syntax.PropertyFetch:
				id, ok := x.Name.(*syntax.Identifier)
				v, isVar := x.Var.(*syntax.Variable)
				if ok && isVar && v.Name == "this" {
					return propWriteKey{id.Value, false}, true
				}
			case *syntax.StaticPropertyFetch:
				if v, ok := x.Name.(*syntax.Variable); ok && pc.staticClass(x, x.Class) == class {
					return propWriteKey{v.Name, true}, true
				}
			}
			return propWriteKey{}, false
		}
		entry := func(k propWriteKey) *propWrites {
			w := idx[k]
			if w == nil {
				w = &propWrites{}
				idx[k] = w
			}
			return w
		}
		for _, m := range class.Members {
			if p, ok := m.(*syntax.Property); ok {
				for _, it := range p.Props {
					if it.Var != nil {
						decls[it.Var.Name] = propDecl{it, p}
					}
				}
			}
			syntax.Inspect(m, func(n syntax.Node) bool {
				switch x := n.(type) {
				case *syntax.Assign:
					if k, ok := keyOf(x.Var); ok {
						w := entry(k)
						if x.Op.Kind != syntax.TEqual || x.ByRef {
							w.dirty = true
						} else {
							w.vals = append(w.vals, AssignedValue(x))
						}
					}
				case *syntax.IncDec:
					if k, ok := keyOf(x.Var); ok {
						entry(k).dirty = true
					}
				}
				return true
			})
		}
		return &classProps{decls, idx}
	}
	return f.Memo(classPropWritesKey{class}, build).(*classProps)
}
