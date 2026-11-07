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

// Extending lists each class that extends or implements fqn, directly or
// not, through any of its declarations and through the names class_alias()
// gives it and its aliases: breadth first, each class once, fqn and its
// aliases left out. The receiver may be written with a leading backslash.
// The list is kept per name until the index changes; callers must not
// modify it.
func (ix *Index) Extending(fqn string) []string {
	k := strings.TrimPrefix(fqn, `\`)
	g := ix.generation()
	if e, ok := ix.extending.Load(k); ok && e.(extendingEntry).gen == g {
		return e.(extendingEntry).classes
	}
	var out []string
	seen := map[string]bool{strings.ToLower(k): true}
	for queue := []string{fqn}; len(queue) > 0; queue = queue[1:] {
		// The names of the class: its own, its aliases and theirs.
		names := []string{queue[0]}
		for i := 0; i < len(names); i++ {
			for _, alias := range ix.AliasesOf(names[i]) {
				if !seen[strings.ToLower(alias)] {
					seen[strings.ToLower(alias)] = true
					names = append(names, alias)
				}
			}
		}
		for _, name := range names {
			for _, sub := range ix.ChildrenAll(name) {
				if folded := strings.ToLower(sub); !seen[folded] {
					seen[folded] = true
					out = append(out, sub)
					queue = append(queue, sub)
				}
			}
		}
	}
	ix.extending.Store(k, extendingEntry{gen: g, classes: out})
	return out
}

type extendingEntry struct {
	gen     uint64
	classes []string
}

// AliasesOf returns the names the project's class_alias() calls give the
// class fqn, across all layers, sorted. A class declared under such a name
// as its parent extends fqn.
func (ix *Index) AliasesOf(fqn string) []string {
	fqn = key(fqn)
	var out []string
	for l := ix; l != nil; l = l.base {
		out = append(out, l.table(&l.aliasedBy, l.buildAliasedBy)[fqn]...)
	}
	sort.Strings(out)
	return out
}

// buildAliasedBy maps each class class_alias() names, folded, to the
// aliases it is given (one entry per declaration).
func (ix *Index) buildAliasedBy() map[string][]string {
	byOriginal := map[string][]string{}
	for alias, targets := range ix.aliases {
		for _, target := range targets {
			k := key(target.original)
			byOriginal[k] = append(byOriginal[k], alias)
		}
	}
	return byOriginal
}
