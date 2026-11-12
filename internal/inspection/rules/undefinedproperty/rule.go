// Package undefinedproperty implements the UndefinedProperty inspection.
package undefinedproperty

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

// New constructs the stateless inspection.
func New() analysis.Rule { return rule{} }

type rule struct{}

func (rule) ID() string { return "UndefinedProperty" }

func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KPropertyFetch, syntax.KStaticPropertyFetch}
}

func (rule) Semantic() {}

// Check reports a property access whose receiver's classes native
// declarations prove, none of which declares the property, nor `__get` or
// `__isset`, anywhere in its hierarchy (traits and `@property` tags
// included), and none of which allows dynamic properties (`stdClass` and
// its subclasses, `#[AllowDynamicProperties]`). An instance property counts
// only where it is read (see read) and when the file writes no property of
// that name, which may create it; a static property counts wherever it is
// accessed, isset(), empty() and `??` aside. A check of the property's
// existence guarding the access keeps it silent.
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	var (
		name    syntax.Node
		member  string
		classes []string
		exact   bool
	)
	switch f := n.(type) {
	case *syntax.PropertyFetch:
		id, ok := f.Name.(*syntax.Identifier)
		if !ok || !read(ctx, f) || written(ctx)[id.Value] {
			return
		}
		name, member, classes = id, id.Value, semanticquery.MemberReceiverClasses(ctx, f.Var)
	default:
		s := n.(*syntax.StaticPropertyFetch)
		v, ok := s.Name.(*syntax.Variable)
		if !ok || v.Name == "" || checked(s) {
			return
		}
		name, member = v, v.Name
		classes, exact = semanticquery.StaticMemberClasses(ctx, s.Class)
	}
	label, ok := semanticquery.MissingMember(ctx, classes, exact, func(c *index.Class) bool {
		return c.Props[member] != nil || c.Methods["__get"] != nil || c.Methods["__isset"] != nil ||
			c.HasAttr("AllowDynamicProperties") || strings.EqualFold(c.FQN, "stdClass")
	})
	if !ok || semanticquery.MemberGuarded(ctx, n.(syntax.Expr), member, true) {
		return
	}
	ctx.ReportNode(name, "Fix the property name: "+label+" declares no $"+member+".")
}

// read reports whether the property fetch f reads the property where PHP
// warns when it is undeclared: not as (the base of an element written
// through) an assignment target, a by-reference binding, an unset(), a
// foreach target or subject iterated by reference, a destructuring target
// (its keys are read), an increment, an argument that may be taken by
// reference, nor under isset(), empty() or `??`.
func read(ctx *analysis.Context, f *syntax.PropertyFetch) bool {
	if silent(f) {
		return false
	}
	// Writing an element of the property creates it.
	var target syntax.Node = f
	for {
		switch p := target.Parent().(type) {
		case *syntax.Paren:
			target = p
			continue
		case *syntax.ArrayDimFetch:
			if p.Var == target {
				target = p
				continue
			}
		}
		break
	}
	switch p := target.Parent().(type) {
	case *syntax.Assign:
		return p.Var != target && !(p.ByRef && p.Value == target)
	case *syntax.IncDec, *syntax.Unset:
		return false
	case *syntax.Foreach:
		// Only iterating by reference writes the subject.
		return p.Expr == target && !p.ByRef
	case *syntax.ArrayItem:
		return p.Key == target || !p.ByRef && !destructured(p)
	case *syntax.Arg:
		return !semanticquery.ArgumentMayWrite(ctx.Types(), p)
	}
	return true
}

// destructured reports whether the array item it belongs to a list() or
// `[…]` assigned to (an assignment or foreach target).
func destructured(it *syntax.ArrayItem) bool {
	var n syntax.Node = it
	for {
		switch p := n.Parent().(type) {
		case *syntax.Array, *syntax.List, *syntax.ArrayItem:
			n = p
			continue
		case *syntax.Assign:
			return p.Var == n
		case *syntax.Foreach:
			return p.Value == n || p.Key == n
		}
		return false
	}
}

// silent reports whether the instance property fetch f is accessed where
// PHP neither warns nor creates it: under isset(), empty() or `??` (see
// checked), or as the base of a property unset() removes, through elements
// or properties of it (`unset($cart->meta->x)`).
func silent(f *syntax.PropertyFetch) bool {
	if checked(f) {
		return true
	}
	var n syntax.Node = f
	below := false
	for {
		switch p := n.Parent().(type) {
		case *syntax.Paren:
			n = p
			continue
		case *syntax.ArrayDimFetch:
			if p.Var == n {
				n = p
				continue
			}
		case *syntax.PropertyFetch:
			if p.Var == n {
				n, below = p, true
				continue
			}
		case *syntax.Unset:
			return below
		}
		return false
	}
}

// checked reports whether the access x, or an element or property access
// rooted at it, is the subject of isset(), empty() or the left operand of
// `??`, which never warn.
func checked(x syntax.Expr) bool {
	var n syntax.Node = x
	for {
		switch p := n.Parent().(type) {
		case *syntax.Paren:
			n = p
			continue
		case *syntax.ArrayDimFetch:
			if p.Var == n {
				n = p
				continue
			}
		case *syntax.PropertyFetch:
			if p.Var == n {
				n = p
				continue
			}
		case *syntax.Isset, *syntax.Empty:
			return true
		case *syntax.Binary:
			return p.Left == n && p.Op.Kind == syntax.TCoalesce
		}
		return false
	}
}

type writtenKey struct{}

// written collects the names of the instance properties the file writes as
// named fetches (see read): a dynamic property such a write creates may be
// read anywhere.
func written(ctx *analysis.Context) map[string]bool {
	return ctx.Shared(writtenKey{}, func() any {
		out := map[string]bool{}
		syntax.InspectFile(ctx.File, func(n syntax.Node) bool {
			if f, ok := n.(*syntax.PropertyFetch); ok {
				if id, ok := f.Name.(*syntax.Identifier); ok && !out[id.Value] && !silent(f) && !read(ctx, f) {
					out[id.Value] = true
				}
			}
			return true
		})
		return out
	}).(map[string]bool)
}
