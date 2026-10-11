package index

import (
	"sort"
	"strings"
)

// Subclasses returns the FQNs of all classes (in this index and its base
// layers) that directly or transitively extend class fqn via `extends`.
// Interface implementations are not followed. Cycle-safe; each FQN once.
func (ix *Index) Subclasses(fqn string) []string {
	var out []string
	seen := map[string]bool{key(fqn): true}
	queue := []string{fqn}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for l := ix; l != nil; l = l.base {
			for _, child := range l.directSubclasses(cur) {
				if k := key(child); !seen[k] {
					seen[k] = true
					out = append(out, child)
					queue = append(queue, child)
				}
			}
		}
	}
	return out
}

// directSubclasses lists classes of this layer only whose parent is fqn.
// The first declaration of each child decides (as for lookups); a child
// removed concurrently since Children was computed is skipped.
func (ix *Index) directSubclasses(fqn string) []string {
	var out []string
	children := ix.Children(fqn)
	defer ix.runlock(ix.rlock())
	for _, c := range children {
		if cs := ix.classes[key(c)]; len(cs) > 0 && cs[0].Parent != "" && key(cs[0].Parent) == key(fqn) {
			out = append(out, c)
		}
	}
	return out
}

// AliasesOf returns the names the project's class_alias() calls give the
// class fqn, across all layers, sorted. A class declared under such a name
// as its parent extends fqn.
func (ix *Index) AliasesOf(fqn string) []string {
	fqn = key(fqn)
	var out []string
	for l := ix; l != nil; l = l.base {
		l.mu.RLock()
		for alias, targets := range l.aliases {
			for _, target := range targets {
				if aliasMatches(target.original, fqn) {
					out = append(out, alias)
				}
			}
		}
		l.mu.RUnlock()
	}
	sort.Strings(out)
	return out
}

// aliasMatches avoids allocating lowercase copies for the usual ASCII names.
// Unicode uses key so matching retains the index's lowercase semantics.
func aliasMatches(original, folded string) bool {
	original = strings.TrimPrefix(original, `\`)
	for i := range len(original) {
		if original[i] >= 128 {
			return key(original) == folded
		}
	}
	if len(original) != len(folded) {
		return false
	}
	for i := range len(original) {
		c := original[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		if c != folded[i] {
			return false
		}
	}
	return true
}
