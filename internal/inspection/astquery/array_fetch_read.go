package astquery

import "custos/internal/php/syntax"

// ArrayFetchRequiresRead distinguishes ordinary reads from guarded probes and
// slot creation. Only the value chain carries an isset/empty/coalesce guard;
// reads inside an index expression still evaluate normally.
func ArrayFetchRequiresRead(fetch *syntax.ArrayDimFetch) bool {
	for child, parent := syntax.Node(fetch), fetch.Parent(); ; child, parent = parent, parent.Parent() {
		switch p := parent.(type) {
		case *syntax.ArrayDimFetch:
			if p.Var != child {
				return true
			}
		case *syntax.Paren:
		case *syntax.ArrayItem:
			if p.Value != child {
				return true
			}
		case *syntax.Array, *syntax.List:
			switch owner := parent.Parent().(type) {
			case *syntax.Assign:
				if owner.Var != parent {
					return true
				}
			case *syntax.Foreach:
				if owner.Key != parent && owner.Value != parent {
					return true
				}
			case *syntax.ArrayItem:
				if owner.Value != parent {
					return true
				}
			default:
				return true
			}
		case *syntax.Isset, *syntax.Empty, *syntax.Unset:
			return false
		case *syntax.Assign:
			return p.Var != child || p.Op.Kind != syntax.TEqual && p.Op.Kind != syntax.TCoalesceEqual
		case *syntax.Foreach:
			return child != p.Key && child != p.Value
		case *syntax.Binary:
			return p.Op.Kind != syntax.TCoalesce || p.Left != child
		default:
			return true
		}
	}
}
