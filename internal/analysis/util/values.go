package util

import (
	"strings"

	"custos/internal/syntax"
)

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
	pv := possibleValues{f: f, seen: map[syntax.Node]bool{}}
	pv.collect(e)
	if pv.unknown {
		return nil, false
	}
	return pv.out, true
}

type possibleValues struct {
	f       *syntax.File
	seen    map[syntax.Node]bool
	out     []syntax.Expr
	unknown bool
}

func (pv *possibleValues) text(n syntax.Node) string {
	s := n.Span()
	return string(pv.f.Src[s.Start:s.End])
}

func (pv *possibleValues) collect(e syntax.Expr) {
	e = UnwrapParens(e)
	if e == nil || pv.seen[e] || pv.unknown {
		return
	}
	pv.seen[e] = true
	switch x := e.(type) {
	case *syntax.Ternary:
		if x.Then != nil {
			pv.collect(x.Then)
		}
		pv.collect(x.Else)
	case *syntax.Binary:
		if x.Op.Kind != syntax.TCoalesce {
			pv.out = append(pv.out, e)
			return
		}
		pv.collect(x.Left)
		pv.collect(x.Right)
	case *syntax.Variable:
		pv.variable(x)
	case *syntax.PropertyFetch:
		pv.property(x)
	case *syntax.ClassConstFetch:
		pv.classConst(x)
	case *syntax.ConstFetch:
		pv.constant(x)
	default:
		pv.out = append(pv.out, e)
	}
}

// assignedValue follows `$a = $b = v` chains to v.
func assignedValue(a *syntax.Assign) syntax.Expr {
	v := a.Value
	for {
		inner, ok := UnwrapParens(v).(*syntax.Assign)
		if !ok || inner.Op.Kind != syntax.TEqual {
			return v
		}
		v = inner.Value
	}
}

func enclosingScope(n syntax.Node) syntax.Node {
	for p := n.Parent(); p != nil; p = p.Parent() {
		switch p.(type) {
		case *syntax.Function, *syntax.Method, *syntax.Closure, *syntax.ArrowFunction:
			return p
		}
	}
	return nil
}

func scopeParts(scope syntax.Node) (params []*syntax.Param, body syntax.Node) {
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
		v, ok := UnwrapParens(e).(*syntax.Variable)
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

// assignmentsIn calls fn for each plain `=` assignment under root.
func assignmentsIn(root syntax.Node, fn func(*syntax.Assign)) {
	if root == nil {
		return
	}
	syntax.Inspect(root, func(n syntax.Node) bool {
		if a, ok := n.(*syntax.Assign); ok && a.Op.Kind == syntax.TEqual && !a.ByRef {
			fn(a)
		}
		return true
	})
}

func (pv *possibleValues) variable(v *syntax.Variable) {
	if v.NameExpr != nil || v.Name == "" {
		return
	}
	scope := enclosingScope(v)
	if scope == nil {
		return
	}
	params, body := scopeParts(scope)
	if UnstableVariable(body, v.Name) {
		pv.unknown = true
		return
	}
	for _, p := range params {
		if p.Var != nil && p.Var.Name == v.Name && p.Default != nil {
			pv.collect(p.Default)
		}
	}
	var vals []syntax.Expr
	assignmentsIn(body, func(a *syntax.Assign) {
		if t, ok := a.Var.(*syntax.Variable); ok && t.NameExpr == nil && t.Name == v.Name {
			vals = append(vals, assignedValue(a))
		}
	})
	for _, x := range vals {
		pv.collect(x)
	}
}

func enclosingClass(n syntax.Node) *syntax.ClassLike {
	for p := n.Parent(); p != nil; p = p.Parent() {
		if c, ok := p.(*syntax.ClassLike); ok {
			return c
		}
	}
	return nil
}

func (pv *possibleValues) property(p *syntax.PropertyFetch) {
	id, ok := p.Name.(*syntax.Identifier)
	if !ok {
		return
	}
	if v, ok := p.Var.(*syntax.Variable); !ok || v.Name != "this" {
		return
	}
	class := enclosingClass(p)
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
	match := func(a *syntax.Assign) {
		if _, ok := a.Var.(*syntax.PropertyFetch); ok && pv.text(a.Var) == target {
			vals = append(vals, assignedValue(a))
		}
	}
	scope := enclosingScope(p)
	if scope != nil {
		_, body := scopeParts(scope)
		assignmentsIn(body, match)
	}
	for _, m := range class.Members {
		if meth, ok := m.(*syntax.Method); ok && meth.Name != nil &&
			strings.EqualFold(meth.Name.Value, "__construct") && syntax.Node(meth) != scope && meth.Body != nil {
			assignmentsIn(meth.Body, match)
		}
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
		class = enclosingClass(c)
	default:
		want := LastSegment(cn.Value)
		syntax.InspectFile(pv.f, func(n syntax.Node) bool {
			if cl, ok := n.(*syntax.ClassLike); ok && cl.Name != nil && strings.EqualFold(cl.Name.Value, want) && class == nil {
				class = cl
			}
			return class == nil
		})
	}
	if class == nil {
		return
	}
	for _, m := range class.Members {
		if cc, ok := m.(*syntax.ClassConst); ok {
			for _, it := range cc.Consts {
				if it.Name != nil && it.Name.Value == id.Value && it.Value != nil {
					pv.collect(it.Value)
				}
			}
		}
	}
}

func (pv *possibleValues) constant(c *syntax.ConstFetch) {
	if c.Name == nil {
		return
	}
	name := LastSegment(c.Name.Value)
	switch strings.ToLower(name) {
	case "true", "false", "null":
		pv.out = append(pv.out, c)
		return
	}
	var found syntax.Expr
	syntax.InspectFile(pv.f, func(n syntax.Node) bool {
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
	if found != nil {
		pv.collect(found)
	}
}

// LastSegment returns the part of a written name after the last backslash.
func LastSegment(written string) string {
	return written[strings.LastIndexByte(written, '\\')+1:]
}
