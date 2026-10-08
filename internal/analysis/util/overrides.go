package util

import (
	"strings"

	"custos/internal/index"
	"custos/internal/phpver"
)

// OverriddenBelow reports whether a class-like extending or implementing
// fqn, at any depth, declares the method lname (lower-case): a call bound
// to fqn's version (`self::m()`) then differs from a dynamic `$this->m()`
// for instances of that descendant. Only indexed descendants are seen.
func OverriddenBelow(ix *index.Index, fqn, lname string, ver phpver.Version) bool {
	seen := map[string]bool{strings.ToLower(strings.TrimPrefix(fqn, `\`)): true}
	queue := []string{fqn}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, ch := range ix.ChildrenAll(cur) {
			k := strings.ToLower(strings.TrimPrefix(ch, `\`))
			if seen[k] {
				continue
			}
			seen[k] = true
			if c := ix.Class(ch, ver); c != nil {
				if _, ok := c.Methods[lname]; ok {
					return true
				}
			}
			queue = append(queue, ch)
		}
	}
	return false
}
