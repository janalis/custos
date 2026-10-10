package flowquery

import "custos/internal/php/syntax"

type nativeStatementsKey struct{}

// NativePriorStatements returns direct preceding statements only when the
// expression belongs to a straight-line statement list. Nested conditionals
// are not flattened. A cached index keeps repeated queries linear overall.
func NativePriorStatements(f *syntax.File, n syntax.Node) []syntax.Stmt {
	var statement syntax.Stmt
	for p := n; p != nil; p = p.Parent() {
		if s, ok := p.(syntax.Stmt); ok {
			statement = s
			break
		}
	}
	if statement == nil {
		return nil
	}
	prior := f.Memo(nativeStatementsKey{}, func() any {
		index := map[syntax.Stmt][]syntax.Stmt{}
		add := func(statements []syntax.Stmt) {
			for i, s := range statements {
				index[s] = statements[:i]
			}
		}
		add(f.Stmts)
		syntax.InspectFile(f, func(node syntax.Node) bool {
			if block, ok := node.(*syntax.Block); ok {
				add(block.Stmts)
			}
			return true
		})
		return index
	}).(map[syntax.Stmt][]syntax.Stmt)
	return prior[statement]
}
