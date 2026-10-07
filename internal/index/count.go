package index

// ClassCount returns how many declarations of the class-like fqn exist in the
// nearest layer declaring it (this index first, then its bases). Layers are
// not summed: a file layered over a project index that already contains the
// same file must not count its classes twice.
func (ix *Index) ClassCount(fqn string) int {
	for l := ix; l != nil; l = l.base {
		l.mu.RLock()
		n := len(l.classes[key(fqn)])
		l.mu.RUnlock()
		if n > 0 {
			return n
		}
	}
	return 0
}
