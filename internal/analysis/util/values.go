package util

import (
	"strings"

	"custos/internal/syntax"
)

// Value discovery: four variants share one traversal (valueWalk) and differ
// only in how they resolve variables, properties and constants, because
// their specs ask for different guarantees:
//
//   - PossibleValues: file-local and permissive (every assignment of a
//     variable in its scope, `$this` properties by source text, constants
//     of classes found by short name);
//   - PossibleValuesReaching: PossibleValues where a variable only takes
//     the assignments that reach the use;
//   - DiscoverValues (discover.go): index-aware (receiver types, inherited
//     members, constants of other files);
//   - PossibleValuesComplete (values_complete.go): conservative, for
//     consumers that need *every* value (anything unaccounted for makes
//     the result incomplete).

// valueWalk collects candidate value expressions, each expression at most
// once, parentheses stripped: both `??` operands, ternary branches,
// anything else through resolve (a value itself when resolve declines).
type valueWalk struct {
	f    *syntax.File
	seen map[syntax.Node]bool
	out  []syntax.Expr
	stop bool // the result is unknown (incomplete): consumers stay silent

	elvisCond bool // `c ?: e` also yields c (otherwise only e)
	nilStops  bool // a missing expression makes the result unknown
	skipEmpty bool // zero-width (recovery) expressions are ignored
	// resolve handles e (variables, properties, constants...); false when
	// e is a value itself.
	resolve func(e syntax.Expr) bool
}

func newValueWalk(f *syntax.File) valueWalk {
	return valueWalk{f: f, seen: map[syntax.Node]bool{}}
}

func (w *valueWalk) result() ([]syntax.Expr, bool) {
	if w.stop {
		return nil, false
	}
	return w.out, true
}

func (w *valueWalk) text(n syntax.Node) string {
	s := n.Span()
	return string(w.f.Src[s.Start:s.End])
}

func (w *valueWalk) collect(e syntax.Expr) {
	e = syntax.UnwrapParens(e)
	if e == nil {
		if w.nilStops {
			w.stop = true
		}
		return
	}
	if w.seen[e] || w.stop || (w.skipEmpty && e.Span().Len() == 0) {
		return
	}
	w.seen[e] = true
	switch x := e.(type) {
	case *syntax.Ternary:
		if x.Then != nil {
			w.collect(x.Then)
		} else if w.elvisCond {
			w.collect(x.Cond)
		}
		w.collect(x.Else)
		return
	case *syntax.Binary:
		if x.Op.Kind == syntax.TCoalesce {
			w.collect(x.Left)
			w.collect(x.Right)
			return
		}
	default:
		if w.resolve(e) {
			return
		}
	}
	w.out = append(w.out, e)
}

// localVar expands a local variable to its parameter default and every
// plain `=` assignment in the enclosing function-like scope (nested
// closures included); nothing in top-level code. An unstable variable or
// more than maxPossibleValues assignments make the result unknown.
func (w *valueWalk) localVar(v *syntax.Variable) {
	if v.NameExpr != nil || v.Name == "" {
		return
	}
	scope := syntax.EnclosingFuncLike(v)
	if scope == nil {
		return
	}
	params, body := ScopeParts(scope)
	ix := assignsUnder(w.f, body)
	if ix.unstable[v.Name] {
		w.stop = true
		return
	}
	for _, p := range params {
		if p.Var != nil && p.Var.Name == v.Name && p.Default != nil {
			w.collect(p.Default)
		}
	}
	assigns := ix.byVar[v.Name]
	if len(assigns) > maxPossibleValues {
		w.stop = true
		return
	}
	for _, a := range assigns {
		w.collect(AssignedValue(a))
	}
}

// PossibleValues collects the candidate value expressions of e, following
// (recursively, each expression at most once, parentheses stripped):
// ternary branches, both `??` operands, local variables (parameter default +
// every plain `=` assignment in the enclosing function-like scope), `$this`
// properties (declared default + assignments in the current scope and the
// constructor), class constants of classes declared in the file and global
// constants defined in the file (`const` / `define`). Anything else is a
// value itself. Unresolvable variables/constants yield nothing.
//
// A variable that is also incremented/decremented or compound-assigned in
// its scope (see UnstableVariable) makes the whole result unknown:
// PossibleValues then returns nil. Consumers for which an empty result is
// not already "no report" must use PossibleValuesKnown.
func PossibleValues(f *syntax.File, e syntax.Expr) []syntax.Expr {
	vals, _ := PossibleValuesKnown(f, e)
	return vals
}

// PossibleValuesKnown is PossibleValues reporting whether the result is
// known; when known is false, vals is nil and consumers must stay silent.
func PossibleValuesKnown(f *syntax.File, e syntax.Expr) (vals []syntax.Expr, known bool) {
	pv := &possibleValues{valueWalk: newValueWalk(f)}
	pv.resolve = pv.resolveExpr
	pv.collect(e)
	return pv.result()
}

type possibleValues struct{ valueWalk }

func (pv *possibleValues) resolveExpr(e syntax.Expr) bool {
	switch x := e.(type) {
	case *syntax.Variable:
		pv.localVar(x)
	case *syntax.PropertyFetch:
		pv.property(x)
	case *syntax.ClassConstFetch:
		pv.classConst(x)
	case *syntax.ConstFetch:
		pv.constant(x)
	default:
		return false
	}
	return true
}

// PossibleValuesReaching is PossibleValuesKnown where a local variable only
// takes the plain `=` assignments that reach the use (ReachingAssignments),
// plus its parameter default when the entry value reaches it; anything
// other than a ternary, `??` or variable is resolved by PossibleValuesKnown.
func PossibleValuesReaching(f *syntax.File, e syntax.Expr) (vals []syntax.Expr, known bool) {
	w := newValueWalk(f)
	w.resolve = func(e syntax.Expr) bool {
		if x, ok := e.(*syntax.Variable); ok {
			w.reachingVar(x)
			return true
		}
		vals, ok := PossibleValuesKnown(f, e)
		if !ok {
			w.stop = true
		} else {
			w.out = append(w.out, vals...)
		}
		return true
	}
	w.collect(e)
	return w.result()
}

func (w *valueWalk) reachingVar(v *syntax.Variable) {
	if v.NameExpr != nil || v.Name == "" {
		return
	}
	scope := syntax.EnclosingFuncLike(v)
	if scope == nil {
		return
	}
	params, body := ScopeParts(scope)
	if UnstableVariableIn(w.f, body, v.Name) {
		w.stop = true
		return
	}
	defs, entry := ReachingAssignmentsIn(w.f, scope, v, v.Name)
	if entry {
		for _, p := range params {
			if p.Var != nil && p.Var.Name == v.Name && p.Default != nil {
				w.collect(p.Default)
			}
		}
	}
	for _, d := range defs {
		w.collect(AssignedValue(d))
	}
}

// AssignedValue follows `$a = $b = v` chains (plain `=` only) to v.
func AssignedValue(a *syntax.Assign) syntax.Expr {
	v := a.Value
	for {
		inner, ok := syntax.UnwrapParens(v).(*syntax.Assign)
		if !ok || inner.Op.Kind != syntax.TEqual {
			return v
		}
		v = inner.Value
	}
}

// ScopeParts returns the parameters and the body (block, or expression for
// an arrow function; nil for an abstract method) of a function-like.
func ScopeParts(scope syntax.Node) (params []*syntax.Param, body syntax.Node) {
	switch s := scope.(type) {
	case *syntax.Function:
		return s.Params, s.Body
	case *syntax.Method:
		if s.Body != nil {
			return s.Params, s.Body
		}
		return s.Params, nil
	case *syntax.Closure:
		return s.Params, s.Body
	case *syntax.ArrowFunction:
		return s.Params, s.Expr
	}
	return nil, nil
}

// UnstableVariable reports whether variable $name is, anywhere under root
// (nested closures included), the operand of `++`/`--` or the target of a
// compound assignment (`+=`, `.=`, `??=`, ...). Value discovery treats such
// a variable's value set as unknown.
func UnstableVariable(root syntax.Node, name string) bool {
	if root == nil || name == "" {
		return false
	}
	found := false
	isVar := func(e syntax.Expr) bool {
		v, ok := syntax.UnwrapParens(e).(*syntax.Variable)
		return ok && v.NameExpr == nil && v.Name == name
	}
	syntax.Inspect(root, func(n syntax.Node) bool {
		if found {
			return false
		}
		switch x := n.(type) {
		case *syntax.IncDec:
			found = isVar(x.Var)
		case *syntax.Assign:
			found = x.Op.Kind != syntax.TEqual && isVar(x.Var)
		}
		return !found
	})
	return found
}

func (pv *possibleValues) property(p *syntax.PropertyFetch) {
	id, ok := p.Name.(*syntax.Identifier)
	if !ok {
		return
	}
	if v, ok := p.Var.(*syntax.Variable); !ok || v.Name != "this" {
		return
	}
	class := syntax.EnclosingClass(p)
	if class == nil {
		return
	}
	for _, m := range class.Members {
		prop, ok := m.(*syntax.Property)
		if !ok {
			continue
		}
		for _, it := range prop.Props {
			if it.Var != nil && it.Var.Name == id.Value && it.Default != nil &&
				!strings.HasSuffix(pv.text(it.Default), id.Value) {
				pv.collect(it.Default)
			}
		}
	}
	target := pv.text(p)
	var vals []syntax.Expr
	add := func(root syntax.Node) {
		ix := assignsUnder(pv.f, root)
		if pv.f == nil { // no text index without the source
			for _, a := range ix.byKind[syntax.KPropertyFetch] {
				if pv.text(a.Var) == target {
					vals = append(vals, AssignedValue(a))
				}
			}
			return
		}
		as := ix.propByText[target]
		if len(vals)+len(as) > maxPossibleValues {
			pv.stop = true
			return
		}
		for _, a := range as {
			vals = append(vals, AssignedValue(a))
		}
	}
	scope := syntax.EnclosingFuncLike(p)
	if scope != nil {
		_, body := ScopeParts(scope)
		add(body)
	}
	for _, m := range class.Members {
		if meth, ok := m.(*syntax.Method); ok && meth.Name != nil &&
			strings.EqualFold(meth.Name.Value, "__construct") && syntax.Node(meth) != scope && meth.Body != nil {
			add(meth.Body)
		}
	}
	if pv.stop || len(vals) > maxPossibleValues {
		pv.stop = true
		return
	}
	for _, x := range vals {
		pv.collect(x)
	}
}

func (pv *possibleValues) classConst(c *syntax.ClassConstFetch) {
	id, ok := c.Name.(*syntax.Identifier)
	if !ok {
		return
	}
	cn, ok := c.Class.(*syntax.Name)
	if !ok {
		return
	}
	var class *syntax.ClassLike
	switch strings.ToLower(cn.Value) {
	case "self", "static":
		class = syntax.EnclosingClass(c)
	default:
		class = fileClassNamed(pv.f, LastNamePart(cn.Value))
	}
	if class == nil {
		return
	}
	for _, v := range classConstValues(class, id.Value) {
		pv.collect(v)
	}
}

func (pv *possibleValues) constant(c *syntax.ConstFetch) {
	if c.Name == nil {
		return
	}
	name := LastNamePart(c.Name.Value)
	switch strings.ToLower(name) {
	case "true", "false", "null":
		pv.out = append(pv.out, c)
		return
	}
	if found := fileConstValue(pv.f, name); found != nil {
		pv.collect(found)
	}
}

// fileClassNamed returns the first class-like declared in f (preorder)
// whose short name is want (case-insensitive), or nil.
func fileClassNamed(f *syntax.File, want string) *syntax.ClassLike {
	var found *syntax.ClassLike
	syntax.InspectFile(f, func(n syntax.Node) bool {
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

// classConstValues returns the value expressions of the constants named
// name declared directly in class, in declaration order.
func classConstValues(class *syntax.ClassLike, name string) []syntax.Expr {
	var out []syntax.Expr
	for _, m := range class.Members {
		if cc, ok := m.(*syntax.ClassConst); ok {
			for _, it := range cc.Consts {
				if it.Name != nil && it.Name.Value == name && it.Value != nil {
					out = append(out, it.Value)
				}
			}
		}
	}
	return out
}

// fileConstValue returns the value expression of the global constant name
// declared in f (`const name = v;` or `define('name', v)`), or nil.
func fileConstValue(f *syntax.File, name string) syntax.Expr {
	var found syntax.Expr
	syntax.InspectFile(f, func(n syntax.Node) bool {
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
	return found
}
