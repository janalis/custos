package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

// dynamicBuiltins are the builtin classes (lower case) whose objects answer
// member names their declarations do not list: XML children, array
// entries read as properties, database row columns, COM automation, and
// RecursiveIteratorIterator, which forwards unknown method calls to the
// inner iterator.
var dynamicBuiltins = map[string]bool{
	"simplexmlelement": true, "arrayobject": true, "arrayiterator": true,
	"pdorow": true, "com": true, "variant": true, "dotnet": true,
	"recursiveiteratoriterator": true,
}

// MemberReceiverClasses lists the classes native declarations prove the
// object x to be (see infer.Env.ProvenTypeOf), null aside: nil unless
// every atom names a class (the proven type of `$this` read in a closure or
// an arrow function is unknown: it may run bound to another object).
func MemberReceiverClasses(ctx *analysis.Context, x syntax.Expr) []string {
	typ := ctx.Types().ProvenTypeOf(x).Without("null")
	classes := typ.Classes()
	if len(classes) == 0 || len(classes) != len(typ.Atoms()) {
		return nil
	}
	return classes
}

// StaticMemberClasses lists the class a static access through class names:
// a class name, or `self`, `static` or `parent` outside traits and
// closures (a trait's methods run in the classes using it, a closure may
// be bound to another scope). Exact tells that the member is looked up in
// that class (a class name, `self`, `parent`) rather than in the called
// class (`static`), which may be a subclass.
func StaticMemberClasses(ctx *analysis.Context, class syntax.Expr) (classes []string, exact bool) {
	name, ok := class.(*syntax.Name)
	if !ok {
		return nil, false
	}
	switch strings.ToLower(name.Value) {
	case "self", "static", "parent":
		switch syntax.EnclosingFuncLike(name).(type) {
		case *syntax.Closure, *syntax.ArrowFunction:
			return nil, false
		}
		if c := syntax.EnclosingClass(name); c == nil || c.ClassKind == syntax.KindTrait {
			return nil, false
		}
	}
	cls := StaticCallClass(ctx, name)
	if cls == "" {
		return nil, false
	}
	return []string{cls}, !strings.EqualFold(name.Value, "static")
}

// MissingMember reports whether classes (FQNs) are all known (see
// memberKnown) and has, applied to every class-like of each one's
// hierarchy, finds the member in none of them, nor, unless the lookup is
// exact, in the hierarchy of any indexed class-like extending or
// implementing one of them: the object may be any of those, and code that
// checked which with PHPDoc or a guard the analysis does not follow may
// use what only a subclass declares. A class carrying a @mixin tag, or a
// builtin answering undeclared names, counts as having the member. It
// returns the classes' names, `|`-separated.
func MissingMember(ctx *analysis.Context, classes []string, exact bool, has func(*index.Class) bool) (string, bool) {
	if len(classes) == 0 {
		return "", false
	}
	ix := ctx.Index()
	label := make([]string, len(classes))
	for i, cls := range classes {
		cls = strings.TrimPrefix(cls, `\`)
		ancestors, ok := memberKnown(ctx, cls)
		if !ok || hasAny(ancestors, has) {
			return "", false
		}
		if !exact {
			for _, sub := range ix.Extending(cls) {
				if hasAny(ix.Ancestors(sub, ctx.PHP), has) {
					return "", false
				}
			}
		}
		label[i] = strings.TrimPrefix(ancestors[0].FQN, `\`)
	}
	return strings.Join(label, "|"), true
}

// hasAny reports whether has finds the member in one of classes, or one of
// them forwards undeclared members (see memberForwarder).
func hasAny(classes []*index.Class, has func(*index.Class) bool) bool {
	for _, c := range classes {
		if has(c) || memberForwarder(c) {
			return true
		}
	}
	return false
}

// memberForwarder reports whether objects of class c may answer member
// names no declaration lists: a @mixin tag, or a builtin of
// dynamicBuiltins.
func memberForwarder(c *index.Class) bool {
	return c.Mixin || dynamicBuiltins[strings.ToLower(strings.TrimPrefix(c.FQN, `\`))]
}

// memberKnown returns the hierarchy of class cls (itself first) when its
// members are all known: cls is a class, interface or enum (no trait)
// available at the target version, its hierarchy resolves completely (see
// HierarchyResolved), and each class-like in it is declared once,
// unconditionally.
func memberKnown(ctx *analysis.Context, cls string) ([]*index.Class, bool) {
	ix := ctx.Index()
	c := ix.Class(cls, ctx.PHP)
	if c == nil || c.Kind == syntax.KindTrait || !ix.AncestorsComplete(cls, ctx.PHP) || !HierarchyResolved(ix, cls, ctx.PHP) {
		return nil, false
	}
	ancestors := ix.Ancestors(cls, ctx.PHP)
	for _, a := range ancestors {
		if a.Conditional || len(ix.ClassDecls(a.FQN, ctx.PHP)) != 1 {
			return nil, false
		}
	}
	return ancestors, true
}

// memberChecks are the builtins whose call asks whether an object or class
// has a member: a method (method_exists, is_callable) or a property
// (property_exists).
var memberChecks = map[string]bool{"method_exists": true, "is_callable": true, "property_exists": true}

// MemberGuarded reports whether the access at may be meant to run only
// when member exists: a condition enclosing it checks a member of that
// name (method_exists(), property_exists(), is_callable() or, for a
// property, isset() of a property of that name), or the conditions guarding
// at narrow the result of such a check made earlier in its scope (an
// early exit, an else arm, a short circuit). Property names compare
// case-sensitively, method names not.
func MemberGuarded(ctx *analysis.Context, at syntax.Expr, member string, property bool) bool {
	for child, p := syntax.Node(at), at.Parent(); p != nil; child, p = p, p.Parent() {
		var conds []syntax.Expr
		switch p := p.(type) {
		case *syntax.If:
			conds = append(conds, p.Cond)
			for _, e := range p.ElseIfs {
				conds = append(conds, e.Cond)
			}
		case *syntax.While:
			conds = append(conds, p.Cond)
		case *syntax.Ternary:
			conds = append(conds, p.Cond)
		case *syntax.Binary:
			if child == p.Right {
				conds = append(conds, p.Left)
			}
		}
		for _, cond := range conds {
			if checksMember(ctx, cond, member, property) {
				return true
			}
		}
	}
	return NativeEarlierCallProves(ctx, at, func(x syntax.Expr) (types.Type, func(types.Type) bool) {
		if !checksMember(ctx, x, member, property) {
			return types.Unknown, nil
		}
		return types.Bool, func(t types.Type) bool { return t.OnlyOf("true") || t.OnlyOf("false") }
	})
}

// checksMember reports whether x calls a member check (see memberChecks)
// given member as a literal string, directly or in an array, or, for a
// property, holds isset() of a property named member.
func checksMember(ctx *analysis.Context, x syntax.Expr, member string, property bool) bool {
	found := false
	syntax.Inspect(x, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.FuncCall:
			if memberChecks[ctx.GlobalFunctionName(n)] {
				syntax.Inspect(n.Args, func(a syntax.Node) bool {
					if lit, ok := a.(*syntax.Literal); ok {
						if s, ok := syntax.LiteralStringValue(lit); ok && sameMember(s, member, property) {
							found = true
						}
					}
					return !found
				})
			}
		case *syntax.Isset:
			if property {
				syntax.Inspect(n, func(a syntax.Node) bool {
					if f, ok := a.(*syntax.PropertyFetch); ok {
						if id, ok := f.Name.(*syntax.Identifier); ok && id.Value == member {
							found = true
						}
					}
					return !found
				})
			}
		}
		return !found
	})
	return found
}

// sameMember compares member names: case-sensitively for properties,
// case-insensitively for methods.
func sameMember(a, b string, property bool) bool {
	if property {
		return a == b
	}
	return strings.EqualFold(a, b)
}
