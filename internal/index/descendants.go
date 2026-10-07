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
func (ix *Index) directSubclasses(fqn string) []string {
	var out []string
	for _, c := range ix.Children(fqn) {
		if cls := ix.layerClass(c); cls != nil && cls.Parent != "" && key(cls.Parent) == key(fqn) {
			out = append(out, c)
		}
	}
	return out
}

// layerClass looks a class up in this layer only.
func (ix *Index) layerClass(fqn string) *Class {
	ix.mu.RLock()
	defer ix.mu.RUnlock()
	if cs := ix.classes[key(fqn)]; len(cs) > 0 {
		return cs[0]
	}
	return nil
}
