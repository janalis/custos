package astquery

import (
	"custos/internal/php/syntax"
)

func InNamedNamespace(n syntax.Node) bool {
	for p := n.Parent(); p != nil; p = p.Parent() {
		if ns, ok := p.(*syntax.Namespace); ok {
			return ns.Name != nil
		}
	}
	return false
}
