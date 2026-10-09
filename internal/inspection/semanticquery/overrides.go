package semanticquery

import (
	"strings"

	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

// DescendantMethods returns the lower-case names of the methods declared by
// any indexed class-like extending or implementing fqn, at any depth: a call
// bound to fqn's version of such a method (`self::m()`) differs from a
// dynamic `$this->m()` for instances of that descendant. The walk visits
// every descendant once, so callers checking several methods of one class
// should compute it once (ctx.Memo): a base class with thousands of
// generated subclasses (Google API models) made a per-method walk
// quadratic. In an (invalid) inheritance cycle fqn is its own descendant.
func DescendantMethods(ix *index.Index, fqn string, ver phpversion.Version) map[string]bool {
	out := map[string]bool{}
	seen := map[string]bool{} // fqn itself is reached only through a cycle, then it overrides itself
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
				for m := range c.Methods {
					out[m] = true
				}
			}
			queue = append(queue, ch)
		}
	}
	return out
}
