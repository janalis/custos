package flowquery

import (
	"custos/internal/php/syntax"
)

// Reachable reports whether node n (inside scope's body) can be reached
// from the scope entry: no statement on the path from n up to scope is
// preceded, in its statement list, by a statement that Terminates. Branches
// guarded by constant conditions are considered reachable. scope may be nil
// (file level).
func Reachable(n, scope syntax.Node) bool {
	child := n
	for p := n.Parent(); p != nil && child != scope; p = p.Parent() {
		if list, ok := syntax.StmtListOf(p); ok {
			// Statements before child (all of them when child is not in the
			// list, e.g. a case condition); FirstTerminating is cached so
			// long lists stay linear overall.
			i := syntax.StmtIndex(list, child)
			if i < 0 {
				i = len(list)
			}
			if syntax.FirstTerminating(p) < i {
				return false
			}
		}
		if p == scope {
			break
		}
		child = p
	}
	return true
}
