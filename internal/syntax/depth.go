package syntax

// TreeDepth returns the maximum node depth of the given statements, computed
// iteratively (no recursion, so it is safe on arbitrarily deep trees). It
// stops early once MaxDepth is exceeded.
func TreeDepth(stmts []Stmt) int {
	type item struct {
		n Node
		d int
	}
	stack := make([]item, 0, 64)
	for _, s := range stmts {
		if s != nil {
			stack = append(stack, item{s, 1})
		}
	}
	maxD := 0
	for len(stack) > 0 {
		it := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if it.d > maxD {
			maxD = it.d
			if maxD > MaxDepth {
				return maxD
			}
		}
		Children(it.n, func(c Node) { stack = append(stack, item{c, it.d + 1}) })
	}
	return maxD
}
