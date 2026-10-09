package flowquery

import (
	"strings"

	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// ValueWalk collects candidate value expressions, each expression at most
// once, parentheses stripped: both `??` operands, ternary branches,
// anything else through resolve (a value itself when resolve declines).
type ValueWalk struct {
	F    *syntax.File
	Seen map[syntax.Node]bool
	Out  []syntax.Expr
	Stop bool // the result is unknown (incomplete): consumers stay silent

	ElvisCond bool // `c ?: e` also yields c (otherwise only e)
	NilStops  bool // a missing expression makes the result unknown
	SkipEmpty bool // zero-width (recovery) expressions are ignored
	// resolve handles e (variables, properties, constants...); false when
	// e is a value itself.
	Resolve func(e syntax.Expr) bool
}

func NewValueWalk(f *syntax.File) ValueWalk {
	return ValueWalk{F: f, Seen: map[syntax.Node]bool{}}
}

func (w *ValueWalk) Result() ([]syntax.Expr, bool) {
	if w.Stop {
		return nil, false
	}
	return w.Out, true
}

func (w *ValueWalk) Text(n syntax.Node) string {
	s := n.Span()
	return string(w.F.Src[s.Start:s.End])
}

func (w *ValueWalk) Collect(e syntax.Expr) {
	e = syntax.UnwrapParens(e)
	if e == nil {
		if w.NilStops {
			w.Stop = true
		}
		return
	}
	if w.Seen[e] || w.Stop || (w.SkipEmpty && e.Span().Len() == 0) {
		return
	}
	w.Seen[e] = true
	switch x := e.(type) {
	case *syntax.Ternary:
		if x.Then != nil {
			w.Collect(x.Then)
		} else if w.ElvisCond {
			w.Collect(x.Cond)
		}
		w.Collect(x.Else)
		return
	case *syntax.Binary:
		if x.Op.Kind == syntax.TCoalesce {
			w.Collect(x.Left)
			w.Collect(x.Right)
			return
		}
	default:
		if w.Resolve(e) {
			return
		}
	}
	w.Out = append(w.Out, e)
}

// LocalVar expands a local variable to its parameter default and every
// plain `=` assignment in the enclosing function-like scope (nested
// closures included); nothing in top-level code. An unstable variable or
// more than maxPossibleValues assignments make the result unknown.
func (w *ValueWalk) LocalVar(v *syntax.Variable) {
	if v.NameExpr != nil || v.Name == "" {
		return
	}
	scope := syntax.EnclosingFuncLike(v)
	if scope == nil {
		return
	}
	params, body := astquery.ScopeParts(scope)
	ix := assignsUnder(w.F, body)
	if ix.unstable[v.Name] {
		w.Stop = true
		return
	}
	for _, p := range params {
		if p.Var != nil && p.Var.Name == v.Name && p.Default != nil {
			w.Collect(p.Default)
		}
	}
	assigns := ix.byVar[v.Name]
	if len(assigns) > MaxPossibleValues {
		w.Stop = true
		return
	}
	for _, a := range assigns {
		w.Collect(astquery.AssignedValue(a))
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
// its scope (see UnstableVariableIn) makes the whole result unknown:
// PossibleValues then returns nil. Consumers for which an empty result is
// not already "no report" must use PossibleValuesKnown.
func PossibleValues(f *syntax.File, e syntax.Expr) []syntax.Expr {
	vals, _ := PossibleValuesKnown(f, e)
	return vals
}

// PossibleValuesKnown is PossibleValues reporting whether the result is
// known; when known is false, vals is nil and consumers must stay silent.
func PossibleValuesKnown(f *syntax.File, e syntax.Expr) (vals []syntax.Expr, known bool) {
	pv := &possibleValues{ValueWalk: NewValueWalk(f)}
	pv.Resolve = pv.resolveExpr
	pv.Collect(e)
	return pv.Result()
}

type possibleValues struct{ ValueWalk }

func (pv *possibleValues) resolveExpr(e syntax.Expr) bool {
	switch x := e.(type) {
	case *syntax.Variable:
		pv.LocalVar(x)
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
	w := NewValueWalk(f)
	w.Resolve = func(e syntax.Expr) bool {
		if x, ok := e.(*syntax.Variable); ok {
			w.reachingVar(x)
			return true
		}
		vals, ok := PossibleValuesKnown(f, e)
		if !ok {
			w.Stop = true
		} else {
			w.Out = append(w.Out, vals...)
		}
		return true
	}
	w.Collect(e)
	return w.Result()
}

func (w *ValueWalk) reachingVar(v *syntax.Variable) {
	if v.NameExpr != nil || v.Name == "" {
		return
	}
	scope := syntax.EnclosingFuncLike(v)
	if scope == nil {
		return
	}
	params, body := astquery.ScopeParts(scope)
	if UnstableVariableIn(w.F, body, v.Name) {
		w.Stop = true
		return
	}
	defs, entry := ReachingAssignmentsIn(w.F, scope, v, v.Name)
	if entry {
		for _, p := range params {
			if p.Var != nil && p.Var.Name == v.Name && p.Default != nil {
				w.Collect(p.Default)
			}
		}
	}
	for _, d := range defs {
		w.Collect(astquery.AssignedValue(d))
	}
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
				!strings.HasSuffix(pv.Text(it.Default), id.Value) {
				pv.Collect(it.Default)
			}
		}
	}
	target := pv.Text(p)
	var vals []syntax.Expr
	add := func(root syntax.Node) {
		as := assignsUnder(pv.F, root).propByText[target]
		if len(vals)+len(as) > MaxPossibleValues {
			pv.Stop = true
			return
		}
		for _, a := range as {
			vals = append(vals, astquery.AssignedValue(a))
		}
	}
	scope := syntax.EnclosingFuncLike(p)
	if scope != nil {
		_, body := astquery.ScopeParts(scope)
		add(body)
	}
	for _, m := range class.Members {
		if meth, ok := m.(*syntax.Method); ok && meth.Name != nil &&
			strings.EqualFold(meth.Name.Value, "__construct") && syntax.Node(meth) != scope && meth.Body != nil {
			add(meth.Body)
		}
	}
	if pv.Stop { // add keeps vals within maxPossibleValues
		return
	}
	for _, x := range vals {
		pv.Collect(x)
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
		class = astquery.FileClassNamed(pv.F, astquery.LastNamePart(cn.Value))
	}
	if class == nil {
		return
	}
	for _, v := range astquery.ClassConstValues(class, id.Value) {
		pv.Collect(v)
	}
}

func (pv *possibleValues) constant(c *syntax.ConstFetch) {
	name := astquery.LastNamePart(c.Name.Value)
	switch strings.ToLower(name) {
	case "true", "false", "null":
		pv.Out = append(pv.Out, c)
		return
	}
	if found := astquery.FileConstValue(pv.F, name); found != nil {
		pv.Collect(found)
	}
}
