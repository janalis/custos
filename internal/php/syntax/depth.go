package syntax

// TreeDepth returns the maximum node depth of the given statements. The
// walk stops descending as soon as MaxDepth is exceeded, so its recursion is
// bounded (MaxDepth+1 levels) and safe on arbitrarily deep trees; unlike an
// explicit stack it allocates nothing per node (an explicit stack of every
// pending sibling cost ~1 GB of allocations on a 3 MB file of 1M top-level
// statements).
func TreeDepth(stmts []Stmt) int {
	maxD := 0
	var visit func(n Node, d int)
	visit = func(n Node, d int) {
		if d > maxD {
			maxD = d
		}
		if maxD > MaxDepth {
			return
		}
		Children(n, func(c Node) {
			if maxD <= MaxDepth {
				visit(c, d+1)
			}
		})
	}
	for _, s := range stmts {
		if maxD > MaxDepth {
			break
		}
		if s != nil && !isNilNode(s) {
			visit(s, 1)
		}
	}
	return maxD
}
