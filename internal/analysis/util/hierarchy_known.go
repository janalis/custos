package util

import (
	"custos/internal/index"
	"custos/internal/phpver"
)

// HierarchyResolved reports whether every parent, interface and trait of
// class fqn (transitively) resolves in the index. When one does not, the
// missing ancestor may declare members (methods, __isset, __toString…) the
// index cannot see, so member-based checks must stay silent.
func HierarchyResolved(ix *index.Index, fqn string, ver phpver.Version) bool {
	for _, c := range ix.Ancestors(fqn, ver) {
		if c.Parent != "" && ix.Class(c.Parent, ver) == nil {
			return false
		}
		for _, r := range c.Interfaces {
			if ix.Class(r, ver) == nil {
				return false
			}
		}
		for _, r := range c.Traits {
			if ix.Class(r, ver) == nil {
				return false
			}
		}
	}
	return true
}
