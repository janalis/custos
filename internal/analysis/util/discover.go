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
	d := &discoverer{valueWalk: newValueWalk(env.File), env: env}
	d.elvisCond, d.skipEmpty = true, true
	d.resolve = d.resolveExpr
	d.collect(e)
	return d.result()
}

type discoverer struct {
	valueWalk
	env *infer.Env
}

func (d *discoverer) resolveExpr(e syntax.Expr) bool {
	switch x := e.(type) {
	case *syntax.Variable:
		d.localVar(x)
	case *syntax.PropertyFetch:
		if id, ok := x.Name.(*syntax.Identifier); ok {
			d.property(x, d.env.TypeOf(x.Var).Classes(), id.Value)
		}
	case *syntax.StaticPropertyFetch:
		if v, ok := x.Name.(*syntax.Variable); ok && v.NameExpr == nil && v.Name != "" {
			if cls := d.env.ClassRef(x.Class); cls != "" {
				d.property(x, []string{cls}, v.Name)
			}
		}
	case *syntax.ClassConstFetch:
		d.classConst(x)
	case *syntax.ConstFetch:
		d.constant(x)
	default:
		return false
	}
	return true
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
	scan := func(root syntax.Node) {
		cands := assignsUnder(d.f, root).byKind[fetch.Kind()]
		if len(cands) > maxAssignScan {
			d.stop = true
			return
		}
		for _, a := range cands {
			if Equivalent(d.f, a.Var, fetch) {
				vals = append(vals, AssignedValue(a))
			}
		}
	}
	scope := syntax.EnclosingFuncLike(fetch)
	if scope != nil {
		_, body := ScopeParts(scope)
		scan(body)
	}
	for _, m := range decl.Members {
		if meth, ok := m.(*syntax.Method); ok && meth.Name != nil && meth.Body != nil &&
			strings.EqualFold(meth.Name.Value, "__construct") && syntax.Node(meth) != scope {
			scan(meth.Body)
		}
	}
	if d.stop {
		return
	}
	if len(vals) > maxPossibleValues {
		d.stop = true
		return
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
	cls := d.env.ClassRef(c.Class)
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
	for _, v := range classConstValues(decl, id.Value) {
		d.collect(v)
	}
}

func (d *discoverer) constant(c *syntax.ConstFetch) {
	if c.Name == nil {
		return
	}
	switch strings.ToLower(LastNamePart(c.Name.Value)) {
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
