package astquery

import (
	"custos/internal/php/syntax"
)

// ClassScope returns the class-like whose scope n runs in: the nearest
// enclosing class-like, unless a named function declaration (which never has
// a class scope) comes first.
func ClassScope(n syntax.Node) *syntax.ClassLike {
	for p := n.Parent(); p != nil; p = p.Parent() {
		switch p := p.(type) {
		case *syntax.Function:
			return nil
		case *syntax.ClassLike:
			return p
		}
	}
	return nil
}
