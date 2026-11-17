// Package readmap provides a concurrent map for caches that store each key
// once and then read it many times from parallel workers.
//
// sync.Map serves the same pattern but boxes keys into interfaces, hashing
// them through their dynamic type on every read. Map keeps its keys typed:
// a read that finds its key is an atomic load and a plain map lookup,
// without a lock or a write to shared memory.
package readmap

import (
	"maps"
	"sync"
	"sync/atomic"
)

// Map is a concurrent map whose entries never change once stored. The zero
// Map is empty and ready to use; a Map must not be copied after first use.
//
// Reads look up an immutable map. Stores go to a locked map holding every
// entry; it replaces the immutable one once reads have missed as many times
// as it has entries, so that its copying stays proportional to the misses.
// Until then, a read missing the immutable map takes the lock: Map suits
// caches that store what they miss, not lookups that mostly find nothing.
type Map[K comparable, V any] struct {
	read atomic.Pointer[view[K, V]]

	mu     sync.Mutex
	dirty  map[K]V // all entries when read lacks some, nil otherwise
	misses int
}

type view[K comparable, V any] struct {
	m map[K]V
	// amended tells that dirty holds entries m lacks.
	amended bool
}

// Load returns the value stored under k.
func (m *Map[K, V]) Load(k K) (V, bool) {
	r := m.read.Load()
	if r == nil {
		var zero V
		return zero, false
	}
	if v, ok := r.m[k]; ok || !r.amended {
		return v, ok
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	// The read map may have been replaced while this read waited for the
	// lock; when it holds every entry, dirty is nil and finds nothing.
	r = m.read.Load()
	v, ok := r.m[k]
	if !ok {
		v, ok = m.dirty[k]
	}
	if r.amended {
		m.misses++
		if m.misses >= len(m.dirty) {
			m.read.Store(&view[K, V]{m: m.dirty})
			m.dirty, m.misses = nil, 0
		}
	}
	return v, ok
}

// Store stores v under k unless k already has a value, and reports whether
// it did. Entries never change: callers store equivalent values for a key.
func (m *Map[K, V]) Store(k K, v V) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	r := m.read.Load()
	if r == nil {
		r = &view[K, V]{}
	}
	if _, ok := r.m[k]; ok {
		return false
	}
	if _, ok := m.dirty[k]; ok {
		return false
	}
	if m.dirty == nil {
		m.dirty = make(map[K]V, len(r.m)+1)
		maps.Copy(m.dirty, r.m)
	}
	m.dirty[k] = v
	if !r.amended {
		m.read.Store(&view[K, V]{m: r.m, amended: true})
	}
	return true
}

// Len returns the number of entries.
func (m *Map[K, V]) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.dirty != nil {
		return len(m.dirty)
	}
	if r := m.read.Load(); r != nil {
		return len(r.m)
	}
	return 0
}
