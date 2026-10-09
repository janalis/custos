package index

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
	ix.mu.RLock()
	defer ix.mu.RUnlock()
	for _, c := range children {
		if cs := ix.classes[key(c)]; len(cs) > 0 && cs[0].Parent != "" && key(cs[0].Parent) == key(fqn) {
			out = append(out, c)
		}
	}
	return out
}
