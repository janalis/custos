package astquery

import (
	"custos/internal/php/syntax"
)

// PatternHasTarget reports whether a destructuring pattern assigns anything
// (recursively through nested patterns).
func PatternHasTarget(items []*syntax.ArrayItem) bool {
	for _, it := range items {
		if it == nil || it.Value == nil {
			continue
		}
		switch v := it.Value.(type) {
		case *syntax.List:
			if PatternHasTarget(v.Items) {
				return true
			}
		case *syntax.Array:
			if PatternHasTarget(v.Items) {
				return true
			}
		default:
			return true
		}
	}
	return false
}
