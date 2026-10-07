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
//     file): only properties declared in the class itself (without hooks);
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
	pc := completeValues{f: f, seen: map[syntax.Node]bool{}}
	pc.collect(e)
	if pc.incomplete {
		return nil, false
	}
	return pc.out, true
}

type completeValues struct {
	f          *syntax.File
	seen       map[syntax.Node]bool
	out        []syntax.Expr
	incomplete bool
}

func (pc *completeValues) collect(e syntax.Expr) {
	e = UnwrapParens(e)
	if e == nil {
		pc.incomplete = true
		return
	}
	if pc.seen[e] || pc.incomplete {
		return
	}
	pc.seen[e] = true
	switch x := e.(type) {
	case *syntax.Ternary:
		if x.Then != nil {
			pc.collect(x.Then)
		} else {
			pc.collect(x.Cond)
		}
		pc.collect(x.Else)
	case *syntax.Binary:
		if x.Op.Kind != syntax.TCoalesce {
			pc.out = append(pc.out, e)
			return
		}
		pc.collect(x.Left)
		pc.collect(x.Right)
	case *syntax.Variable:
		pc.variable(x)
	case *syntax.PropertyFetch:
		id, ok := x.Name.(*syntax.Identifier)
		v, isVar := x.Var.(*syntax.Variable)
		if !ok || !isVar || v.Name != "this" || x.NullSafe {
			pc.incomplete = true
			return
		}
		pc.property(enclosingClass(x), id.Value, false)
	case *syntax.StaticPropertyFetch:
		v, ok := x.Name.(*syntax.Variable)
		if !ok || v.NameExpr != nil || v.Name == "" {
			pc.incomplete = true
			return
		}
		pc.property(pc.staticClass(x, x.Class), v.Name, true)
	case *syntax.ClassConstFetch:
		pc.classConst(x)
	case *syntax.ConstFetch:
		pc.constant(x)
	default:
		pc.out = append(pc.out, e)
	}
}

func (pc *completeValues) variable(v *syntax.Variable) {
	if v.NameExpr != nil || v.Name == "" || v.Name == "this" {
		pc.incomplete = true
		return
	}
	scope := enclosingScope(v)
	if scope == nil {
		pc.incomplete = true
		return
	}
	_, body := scopeParts(scope)
	if body == nil {
		pc.incomplete = true
		return
	}
	ix := assignsUnder(pc.f, body)
	if otherWrites(pc.f, scope)[v.Name] || ix.byRefUse[v.Name] || ix.unstable[v.Name] || len(ix.byVar[v.Name]) > maxPossibleValues {
		pc.incomplete = true
		return
	}
	defs, entry := ReachingAssignmentsIn(pc.f, scope, v, v.Name)
	if entry || len(defs) == 0 {
		pc.incomplete = true
		return
	}
	for _, d := range defs {
		pc.collect(assignedValue(d))
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
		return enclosingClass(at)
	case "static":
		if c := enclosingClass(at); c != nil && c.Modifiers.Has(syntax.TFinal) {
			return c
		}
		return nil
	case "parent":
		return nil
	}
	want := LastSegment(cn.Value)
	var found *syntax.ClassLike
	syntax.InspectFile(pc.f, func(n syntax.Node) bool {
		if found != nil {
			return false
		}
		if cl, ok := n.(*syntax.ClassLike); ok && cl.Name != nil && strings.EqualFold(cl.Name.Value, want) {
			found = cl
		}
		return true
	})
	return found
}

func (pc *completeValues) property(class *syntax.ClassLike, name string, static bool) {
	if class == nil {
		pc.incomplete = true
		return
	}
	var decl *syntax.PropertyItem
	var prop *syntax.Property
	for _, m := range class.Members {
		if p, ok := m.(*syntax.Property); ok {
			for _, it := range p.Props {
				if it.Var != nil && it.Var.Name == name {
					decl, prop = it, p
				}
			}
		}
	}
	if decl == nil || prop.Modifiers.Has(syntax.TStatic) != static || len(prop.Hooks) > 0 {
		pc.incomplete = true
		return
	}
	switch {
	case decl.Default != nil:
		pc.collect(decl.Default)
	case prop.Type == nil:
		pc.incomplete = true // implicit null
		return
	}
	matches := func(e syntax.Expr) bool {
		switch x := UnwrapParens(e).(type) {
		case *syntax.PropertyFetch:
			id, ok := x.Name.(*syntax.Identifier)
			v, isVar := x.Var.(*syntax.Variable)
			return !static && ok && isVar && v.Name == "this" && id.Value == name
		case *syntax.StaticPropertyFetch:
			v, ok := x.Name.(*syntax.Variable)
			return static && ok && v.Name == name && pc.staticClass(x, x.Class) == class
		}
		return false
	}
	var vals []syntax.Expr
	for _, m := range class.Members {
		syntax.Inspect(m, func(n syntax.Node) bool {
			switch x := n.(type) {
			case *syntax.Assign:
				if matches(x.Var) {
					if x.Op.Kind != syntax.TEqual || x.ByRef {
						pc.incomplete = true
					} else {
						vals = append(vals, assignedValue(x))
					}
				}
			case *syntax.IncDec:
				if matches(x.Var) {
					pc.incomplete = true
				}
			}
			return !pc.incomplete
		})
	}
	if pc.incomplete {
		return
	}
	if decl.Default == nil && len(vals) == 0 {
		pc.incomplete = true
		return
	}
	for _, v := range vals {
		pc.collect(v)
	}
}

func (pc *completeValues) classConst(c *syntax.ClassConstFetch) {
	id, ok := c.Name.(*syntax.Identifier)
	class := pc.staticClass(c, c.Class)
	if !ok || class == nil {
		pc.incomplete = true
		return
	}
	for _, m := range class.Members {
		if cc, ok := m.(*syntax.ClassConst); ok {
			for _, it := range cc.Consts {
				if it.Name != nil && it.Name.Value == id.Value && it.Value != nil {
					pc.collect(it.Value)
					return
				}
			}
		}
	}
	pc.incomplete = true
}

func (pc *completeValues) constant(c *syntax.ConstFetch) {
	if c.Name == nil {
		pc.incomplete = true
		return
	}
	name := LastSegment(c.Name.Value)
	switch strings.ToLower(name) {
	case "true", "false", "null":
		pc.out = append(pc.out, c)
		return
	}
	var found syntax.Expr
	syntax.InspectFile(pc.f, func(n syntax.Node) bool {
		if found != nil {
			return false
		}
		switch x := n.(type) {
		case *syntax.ConstStmt:
			for _, it := range x.Consts {
				if it.Name != nil && it.Name.Value == name && it.Value != nil {
					found = it.Value
				}
			}
		case *syntax.FuncCall:
			if CallLastName(x) != "define" {
				return true
			}
			args, ok := CallArgValues(x)
			if !ok || len(args) < 2 {
				return true
			}
			if lit, ok := args[0].(*syntax.Literal); ok && lit.LitKind == syntax.LitString {
				if v, ok := StringLiteralValue(lit.Raw); ok && strings.TrimPrefix(v, `\`) == name {
					found = args[1]
				}
			}
		}
		return true
	})
	if found == nil {
		pc.incomplete = true
		return
	}
	pc.collect(found)
}

type otherWritesKey struct{ scope syntax.Node }

// otherWrites returns the names of scope having a write other than a plain
// `$v = …` assignment (computed once per scope and file).
func otherWrites(f *syntax.File, scope syntax.Node) map[string]bool {
	build := func() any {
		m := map[string]bool{}
		for name, accs := range VarAccessesByName(f, scope) {
			for _, acc := range accs {
				if !acc.Write {
					continue
				}
				a, ok := acc.By.(*syntax.Assign)
				if !ok || a.Op.Kind != syntax.TEqual || a.ByRef || UnwrapParens(a.Var) != syntax.Expr(acc.Var) {
					m[name] = true
					break
				}
			}
		}
		return m
	}
	if f == nil {
		return build().(map[string]bool)
	}
	return f.Memo(otherWritesKey{scope}, build).(map[string]bool)
}
