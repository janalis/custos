package flow

import "maps"

// flat returns the snapshot with its own maps, merging an overlay.
func (s *Snapshot) flat() *Snapshot {
	if s.base == nil {
		return s
	}
	n := NewSnapshotWithOptions(s.options)
	for path, file := range s.base.files {
		n.files[path] = file
	}
	n.files[s.own.Path] = s.own
	n.rebuild()
	return n
}

// Replace returns the snapshot NewSnapshotWithOptions would build from its
// files with each file of prev replaced by the file of next at the same
// position (nil removes it, as a file with other options does). Only the
// names of these files are counted again: a fixed point that summarizes a
// few files per round does not re-add the contracts of every other file.
func (s *Snapshot) Replace(prev, next []*File) *Snapshot {
	if s == nil {
		s = NewSnapshotWithOptions(Options{})
	}
	f := s.flat()
	n := &Snapshot{options: f.options, files: maps.Clone(f.files), summaries: maps.Clone(f.summaries), shared: maps.Clone(f.shared)}
	// A name left to a single file whose summary is not at hand yet.
	pending := map[string]bool{}
	count := func(name string) int {
		if _, ok := n.summaries[name]; ok || pending[name] {
			return 1
		}
		return n.shared[name]
	}
	drop := func(old *File) {
		delete(n.files, old.Path)
		for name := range old.Summaries {
			switch c := count(name); c {
			case 1:
				delete(n.summaries, name)
				delete(pending, name)
			case 2:
				delete(n.shared, name)
				pending[name] = true
			default:
				n.shared[name] = c - 1
			}
		}
	}
	for _, old := range prev {
		if old != nil && n.files[old.Path] == old {
			drop(old)
		}
	}
	for _, file := range next {
		if file == nil {
			continue
		}
		if old := n.files[file.Path]; old != nil {
			drop(old)
		}
		if file.options != n.options {
			continue
		}
		n.files[file.Path] = file
		for name, summary := range file.Summaries {
			switch c := count(name); c {
			case 0:
				n.summaries[name] = &summary
			case 1:
				delete(n.summaries, name)
				delete(pending, name)
				if n.shared == nil {
					n.shared = map[string]int{}
				}
				n.shared[name] = 2
			default:
				n.shared[name] = c + 1
			}
		}
	}
	for name := range pending {
		for _, file := range n.files {
			if summary, ok := file.Summaries[name]; ok {
				n.summaries[name] = &summary
			}
		}
	}
	return n
}

func (s *Snapshot) rebuild() {
	for _, f := range s.files {
		for name, summary := range f.Summaries {
			if _, exists := s.summaries[name]; exists || s.shared[name] > 0 {
				if s.shared == nil {
					s.shared = map[string]int{}
				}
				if exists {
					delete(s.summaries, name)
					s.shared[name] = 1
				}
				s.shared[name]++
				continue
			}
			s.summaries[name] = &summary
		}
	}
}
