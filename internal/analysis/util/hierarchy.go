package util

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/index"
	"custos/internal/phpdoc"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// InTestContext reports whether n is in a test context: a test file
// (Context.IsTestFile) or the nearest enclosing class-like is a named class
// whose FQN looks like a test class (IsTestClassFQN). An anonymous class
// ends the search (it has no FQN).
func InTestContext(ctx *analysis.Context, n syntax.Node) bool {
	return ctx.IsTestFile() || IsTestClassFQN(ctx.Names().DeclFQN(syntax.EnclosingClass(n)))
}

// IsTestClassFQN reports whether a class FQN (with or without leading
// backslash) denotes a test class: it ends with Test or contains a \Tests\
// or \Test\ namespace segment.
func IsTestClassFQN(fqn string) bool {
	if fqn == "" {
		return false
	}
	if !strings.HasPrefix(fqn, `\`) {
		fqn = `\` + fqn
	}
	return strings.HasSuffix(fqn, "Test") || strings.Contains(fqn, `\Tests\`) || strings.Contains(fqn, `\Test\`)
}

// ClassDecl returns the declaration node of an indexed class when it lives
// in file f (nil otherwise).
func ClassDecl(f *syntax.File, c *index.Class) *syntax.ClassLike {
	if c == nil || f == nil || c.File != f.Path || c.Span.Len() == 0 {
		return nil
	}
	var out *syntax.ClassLike
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if out != nil {
			return false
		}
		if !n.Span().Contains(c.Span) {
			return false
		}
		if cl, ok := n.(*syntax.ClassLike); ok && cl.Span() == c.Span {
			out = cl
			return false
		}
		return true
	})
	return out
}

// MethodDecl returns the declaration node of an indexed method when its
// class lives in file f (nil otherwise).
func MethodDecl(f *syntax.File, ix *index.Index, m *index.Method, ver phpver.Version) *syntax.Method {
	if m == nil || m.Span.Len() == 0 {
		return nil
	}
	cl := ClassDecl(f, ix.Class(m.Class, ver))
	if cl == nil {
		return nil
	}
	for _, mem := range cl.Members {
		if md, ok := mem.(*syntax.Method); ok && md.Span() == m.Span {
			return md
		}
	}
	return nil
}

// chainWalk visits the class fqn, then (recursively) the traits it uses,
// then its parent class and so on; interfaces are not visited. Each class
// is visited once (cycle-safe), at most index.MaxAncestors classes in all.
// visit returns false to stop.
func chainWalk(ix *index.Index, fqn string, ver phpver.Version, visit func(*index.Class) bool) {
	seen := map[string]bool{}
	budget := index.MaxAncestors
	visit0 := visit
	visit = func(c *index.Class) bool {
		budget--
		return budget >= 0 && visit0(c)
	}
	var walkTraits func(c *index.Class) bool
	walkTraits = func(c *index.Class) bool {
		for _, t := range c.Traits {
			k := strings.ToLower(strings.TrimPrefix(t, `\`))
			if seen[k] {
				continue
			}
			seen[k] = true
			tc := ix.Class(t, ver)
			if tc == nil {
				continue
			}
			if !visit(tc) || !walkTraits(tc) {
				return false
			}
		}
		return true
	}
	for cur := fqn; cur != ""; {
		k := strings.ToLower(strings.TrimPrefix(cur, `\`))
		if seen[k] {
			return
		}
		seen[k] = true
		c := ix.Class(cur, ver)
		if c == nil {
			return
		}
		if !visit(c) || !walkTraits(c) {
			return
		}
		cur = c.Parent
	}
}

// PropertyInChain finds a real (non @property) property declared by the
// class fqn, its ancestors or any trait used along the chain.
func PropertyInChain(ix *index.Index, fqn, name string, ver phpver.Version) *index.Property {
	var out *index.Property
	chainWalk(ix, fqn, ver, func(c *index.Class) bool {
		if p, ok := c.Props[name]; ok && !p.Magic {
			out = p
			return false
		}
		return true
	})
	return out
}

// MethodInChain finds a real (non @method) method (case-insensitive) on the
// class fqn, its ancestors or any trait used along the chain; interface
// methods are not considered.
func MethodInChain(ix *index.Index, fqn, name string, ver phpver.Version) *index.Method {
	lname := strings.ToLower(name)
	var out *index.Method
	chainWalk(ix, fqn, ver, func(c *index.Class) bool {
		if c.Kind == syntax.KindInterface {
			return true
		}
		if m, ok := c.Methods[lname]; ok && m.Span.Len() > 0 && (ver == 0 || m.Avail.In(ver)) {
			out = m
			return false
		}
		return true
	})
	return out
}

// DocHasTag reports whether the doc comment attached to declaration n has a
// tag with the given name (without '@').
func DocHasTag(f *syntax.File, n syntax.Node, tag string) bool {
	c := index.DocComment(f, n)
	return c != "" && phpdoc.Parse(c).Has(tag)
}

// DocHasAnnotation reports whether the doc comment attached to declaration
// n has a tag whose name is not all lower-case (`@Inject`, `@ORM\Column`):
// a framework annotation rather than a phpDoc tag.
func DocHasAnnotation(f *syntax.File, n syntax.Node) bool {
	c := index.DocComment(f, n)
	if c == "" {
		return false
	}
	for _, t := range phpdoc.Parse(c).Tags {
		if t.Name != strings.ToLower(t.Name) {
			return true
		}
	}
	return false
}
