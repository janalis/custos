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

// linkParsedParents validates depth while linking parents and widening spans.
// Invalid trees are discarded by Parse before any consumer can observe them.
func linkParsedParents(f *File) bool {
	f.nodeBase = nodeBase{span: Span{End: uint32(len(f.Src))}}
	var link func(Node, int) bool
	link = func(parent Node, depth int) bool {
		if depth > MaxDepth {
			return false
		}
		valid := true
		Children(parent, func(child Node) {
			if !valid {
				return
			}
			child.base().parent = parent
			valid = link(child, depth+1)
			if child.Span().Start < parent.base().span.Start {
				parent.base().span.Start = child.Span().Start
			}
		})
		return valid
	}
	return link(f, 0)
}
