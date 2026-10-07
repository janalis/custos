package index

import "custos/internal/phpver"

// ClassDecls returns every declaration of the class-like fqn (matched
// case-insensitively) available at ver (0 = any), across all layers. A file
// indexed in an upper layer shadows the same file path in lower layers, so a
// file re-indexed on top of the project index is not counted twice.
func (ix *Index) ClassDecls(fqn string, ver phpver.Version) []*Class {
	var out []*Class
	k := key(fqn)
	for l := ix; l != nil; l = l.base {
		l.mu.RLock()
		for _, c := range l.classes[k] {
			if c.File != "" && shadowedAbove(ix, l, c.File) {
				continue
			}
			if ver == 0 || c.Avail.In(ver) {
				out = append(out, c)
			}
		}
		l.mu.RUnlock()
	}
	return out
}

// shadowedAbove reports whether a layer above `layer` (starting at top)
// indexes path.
func shadowedAbove(top, layer *Index, path string) bool {
	for u := top; u != nil && u != layer; u = u.base {
		u.mu.RLock()
		_, ok := u.files[path]
		u.mu.RUnlock()
		if ok {
			return true
		}
	}
	return false
}

// ChildrenAll returns the FQNs of class-likes directly extending or
// implementing fqn across all layers (unlike Children, which only sees the
// top layer), each once.
func (ix *Index) ChildrenAll(fqn string) []string {
	var out []string
	seen := map[string]bool{}
	for l := ix; l != nil; l = l.base {
		for _, c := range l.Children(fqn) {
			if k := key(c); !seen[k] {
				seen[k] = true
				out = append(out, c)
			}
		}
	}
	return out
}
