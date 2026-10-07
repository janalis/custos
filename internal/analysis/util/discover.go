package util

import (
	"strings"

	"custos/internal/index"
	"custos/internal/infer"
	"custos/internal/syntax"
)

// DiscoverValues is the index-aware variant of PossibleValues (the "value
// discovery" procedure of the CallableMethodValidity spec). It collects the
// candidate value expressions of e (each expression visited once,
// parentheses stripped):
//
//   - ternary: both branches (`a ?: b`: a and b); `a ?? b`: both operands;
//   - variable: parameter default + every plain `=` assignment (innermost
//     value of a chain) in the enclosing function-like scope, nested
//     closures included; nothing in top-level code; unknown result when
//     the variable is unstable (see below);
//   - property fetch (`$o->p`, `X::$p`): resolved through the receiver type
//     and the index; the declared default (unless its text ends with the
//     property name) plus assignments to an equivalent fetch in the current
//     scope and in the declaring class's constructor (declaration in this
//     file only);
//   - class constant: resolved through the index (inheritance aware), its
//     value expression is discovered (declaration in this file only);
//   - global constant other than true/false/null: resolved through the
//     index; a define()/const declared in this file contributes its value
//     expression as-is (not further discovered), one declared elsewhere
//     (project, stubs) contributes the constant reference itself, an
//     unresolved one nothing;
//   - anything else is a value itself.
//
// A variable expanded at any depth that is also incremented/decremented or
// compound-assigned in its scope (UnstableVariable) makes the whole result
// unknown: DiscoverValues then returns nil. Consumers for which an empty
// result is not already "no report" must use DiscoverValuesKnown.
func DiscoverValues(env *infer.Env, e syntax.Expr) []syntax.Expr {
	vals, _ := DiscoverValuesKnown(env, e)
	return vals
}

// DiscoverValuesKnown is DiscoverValues reporting whether the result is
// known; when known is false, vals is nil and consumers must stay silent.
func DiscoverValuesKnown(env *infer.Env, e syntax.Expr) (vals []syntax.Expr, known bool) {
	d := discoverer{env: env, f: env.File, seen: map[syntax.Node]bool{}}
	d.collect(e)
	if d.unknown {
		return nil, false
	}
	return d.out, true
}

type discoverer struct {
	env     *infer.Env
	f       *syntax.File
	seen    map[syntax.Node]bool
	out     []syntax.Expr
	unknown bool
}

func (d *discoverer) text(n syntax.Node) string {
	s := n.Span()
	return string(d.f.Src[s.Start:s.End])
}

func (d *discoverer) collect(e syntax.Expr) {
	e = UnwrapParens(e)
	if e == nil || d.seen[e] || d.unknown || e.Span().Len() == 0 {
		return
	}
	d.seen[e] = true
	switch x := e.(type) {
	case *syntax.Ternary:
		if x.Then != nil {
			d.collect(x.Then)
		} else {
			d.collect(x.Cond)
		}
		d.collect(x.Else)
	case *syntax.Binary:
		if x.Op.Kind != syntax.TCoalesce {
			d.out = append(d.out, e)
			return
		}
		d.collect(x.Left)
		d.collect(x.Right)
	case *syntax.Variable:
		d.variable(x)
	case *syntax.PropertyFetch:
		if id, ok := x.Name.(*syntax.Identifier); ok {
			d.property(x, d.env.TypeOf(x.Var).Classes(), id.Value)
		}
	case *syntax.StaticPropertyFetch:
		if v, ok := x.Name.(*syntax.Variable); ok && v.NameExpr == nil && v.Name != "" {
			if cls := d.classRef(x.Class); cls != "" {
				d.property(x, []string{cls}, v.Name)
			}
		}
	case *syntax.ClassConstFetch:
		d.classConst(x)
	case *syntax.ConstFetch:
		d.constant(x)
	default:
		d.out = append(d.out, e)
	}
}

func (d *discoverer) variable(v *syntax.Variable) {
	if v.NameExpr != nil || v.Name == "" {
		return
	}
	scope := enclosingScope(v)
	if scope == nil {
		return
	}
	params, body := scopeParts(scope)
	if UnstableVariable(body, v.Name) {
		d.unknown = true
		return
	}
	for _, p := range params {
		if p.Var != nil && p.Var.Name == v.Name && p.Default != nil {
			d.collect(p.Default)
		}
	}
	var vals []syntax.Expr
	assignmentsIn(body, func(a *syntax.Assign) {
		if t, ok := a.Var.(*syntax.Variable); ok && t.NameExpr == nil && t.Name == v.Name {
			vals = append(vals, assignedValue(a))
		}
	})
	for _, x := range vals {
		d.collect(x)
	}
}

// classRef resolves a class reference (name or expression) to an FQN
// without leading backslash.
func (d *discoverer) classRef(x syntax.Expr) string {
	if n, ok := x.(*syntax.Name); ok {
		switch strings.ToLower(n.Value) {
		case "self", "static":
			return ClassDeclFQN(d.env.Names, enclosingClass(n))
		case "parent":
			return ParentFQN(d.env.Names, enclosingClass(n))
		}
		return d.env.Names.Class(n.Value, n.Span().Start)
	}
	if cs := d.env.TypeOf(x).Classes(); len(cs) == 1 {
		return strings.TrimPrefix(cs[0], `\`)
	}
	return ""
}

func (d *discoverer) property(fetch syntax.Expr, classes []string, name string) {
	ix, ver := d.env.Index, d.env.PHP
	var decl *syntax.ClassLike
	for _, cls := range classes {
		p := ix.FindProperty(strings.TrimPrefix(cls, `\`), name, ver)
		if p == nil || p.Magic {
			continue
		}
		decl = ClassDecl(d.f, ix.Class(p.Class, ver))
		if decl != nil {
			break
		}
	}
	if decl == nil {
		return
	}
	for _, m := range decl.Members {
		prop, ok := m.(*syntax.Property)
		if !ok {
			continue
		}
		for _, it := range prop.Props {
			if it.Var != nil && it.Var.Name == name && it.Default != nil &&
				!strings.HasSuffix(d.text(it.Default), name) {
				d.collect(it.Default)
			}
		}
	}
	var vals []syntax.Expr
	match := func(a *syntax.Assign) {
		if a.Var.Kind() == fetch.Kind() && Equivalent(d.f, a.Var, fetch) {
			vals = append(vals, assignedValue(a))
		}
	}
	scope := enclosingScope(fetch)
	if scope != nil {
		_, body := scopeParts(scope)
		assignmentsIn(body, match)
	}
	for _, m := range decl.Members {
		if meth, ok := m.(*syntax.Method); ok && meth.Name != nil && meth.Body != nil &&
			strings.EqualFold(meth.Name.Value, "__construct") && syntax.Node(meth) != scope {
			assignmentsIn(meth.Body, match)
		}
	}
	for _, x := range vals {
		d.collect(x)
	}
}

func (d *discoverer) classConst(c *syntax.ClassConstFetch) {
	id, ok := c.Name.(*syntax.Identifier)
	if !ok || strings.EqualFold(id.Value, "class") {
		if !ok {
			return
		}
		d.out = append(d.out, c)
		return
	}
	cls := d.classRef(c.Class)
	if cls == "" {
		return
	}
	ix, ver := d.env.Index, d.env.PHP
	k := ix.FindConst(cls, id.Value, ver)
	if k == nil || k.Case {
		return
	}
	decl := ClassDecl(d.f, ix.Class(k.Class, ver))
	if decl == nil {
		return
	}
	for _, m := range decl.Members {
		if cc, ok := m.(*syntax.ClassConst); ok {
			for _, it := range cc.Consts {
				if it.Name != nil && it.Name.Value == id.Value && it.Value != nil {
					d.collect(it.Value)
				}
			}
		}
	}
}

func (d *discoverer) constant(c *syntax.ConstFetch) {
	if c.Name == nil {
		return
	}
	switch strings.ToLower(LastSegment(c.Name.Value)) {
	case "true", "false", "null":
		d.out = append(d.out, c)
		return
	}
	k := ResolveConstant(d.env, c)
	if k == nil {
		return
	}
	if v := ConstantDeclValue(d.f, k); v != nil {
		d.out = append(d.out, v)
		return
	}
	if k.File != d.f.Path {
		d.out = append(d.out, c)
	}
}

// ResolveConstant resolves a global constant reference through the index
// (namespaced candidate first, then the global fallback); nil when unknown.
func ResolveConstant(env *infer.Env, c *syntax.ConstFetch) *index.Constant {
	if c == nil || c.Name == nil {
		return nil
	}
	fqn, fb := env.Names.Const(c.Name.Value, c.Name.Span().Start)
	k := env.Index.Constant(fqn, env.PHP)
	if k == nil && fb != "" {
		k = env.Index.Constant(fb, env.PHP)
	}
	return k
}

// ConstantDeclValue returns the value expression of an indexed global
// constant declared in file f (`const C = v;` or `define('C', v)`), nil when
// declared elsewhere.
func ConstantDeclValue(f *syntax.File, k *index.Constant) syntax.Expr {
	if k == nil || f == nil || k.File != f.Path || k.Span.Len() == 0 {
		return nil
	}
	var out syntax.Expr
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if out != nil || !n.Span().Contains(k.Span) {
			return false
		}
		if n.Span() != k.Span {
			return true
		}
		switch x := n.(type) {
		case *syntax.ConstItem:
			out = x.Value
		case *syntax.FuncCall:
			if x.Args != nil && len(x.Args.Args) >= 2 {
				if a, ok := x.Args.Args[1].(*syntax.Arg); ok {
					out = a.Value
				}
			}
		}
		return out == nil
	})
	return out
}

// FunctionDecl returns the declaration node of an indexed function when it
// lives in file f (nil otherwise).
func FunctionDecl(f *syntax.File, fn *index.Function) *syntax.Function {
	if fn == nil || f == nil || fn.File != f.Path || fn.Span.Len() == 0 {
		return nil
	}
	var out *syntax.Function
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if out != nil || !n.Span().Contains(fn.Span) {
			return false
		}
		if d, ok := n.(*syntax.Function); ok && d.Span() == fn.Span {
			out = d
			return false
		}
		return true
	})
	return out
}
