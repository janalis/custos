package util

import "custos/internal/syntax"

// Terminates reports whether control never falls through s to the next
// statement of its list (see syntax.Terminates).
func Terminates(s syntax.Stmt) bool { return syntax.Terminates(s) }

// Reachable reports whether node n (inside scope's body) can be reached
// from the scope entry: no statement on the path from n up to scope is
// preceded, in its statement list, by a statement that Terminates. Branches
// guarded by constant conditions are considered reachable. scope may be nil
// (file level).
func Reachable(n, scope syntax.Node) bool {
	var child syntax.Node = n
	for p := n.Parent(); p != nil && child != scope; p = p.Parent() {
		var list []syntax.Stmt
		switch x := p.(type) {
		case *syntax.Block:
			list = x.Stmts
		case *syntax.Case:
			list = x.Stmts
		case *syntax.Namespace:
			list = x.Stmts
		}
		for _, s := range list {
			if syntax.Node(s) == child {
				break
			}
			if Terminates(s) {
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
