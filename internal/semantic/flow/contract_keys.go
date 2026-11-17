package flow

import "math/bits"

// contractKeys numbers the contract keys an environment reads, so that the
// sets of contracts each result read are bit sets instead of string sets.
// An environment shares its table with the workers and the rebased
// environments that reuse its results; none of them runs concurrently with
// another. Numbers never leave the environments sharing the table.
type contractKeys struct {
	ids  map[string]uint32
	keys []string
}

// id returns the number of key, giving it one on its first use.
func (t *contractKeys) id(key string) uint32 {
	if id, ok := t.ids[key]; ok {
		return id
	}
	if t.ids == nil {
		t.ids = map[string]uint32{}
	}
	id := uint32(len(t.keys))
	t.ids[key] = id
	t.keys = append(t.keys, key)
	return id
}

// known returns the number of key when it has one.
func (t *contractKeys) known(key string) (uint32, bool) {
	id, ok := t.ids[key]
	return id, ok
}

// key returns the key numbered id.
func (t *contractKeys) key(id uint32) string { return t.keys[id] }

// keyTable returns the table of the environment, made on first use.
func (s *summaryResolver) keyTable() *contractKeys {
	if s.keys == nil {
		s.keys = &contractKeys{}
	}
	return s.keys
}

// readSet is a set of contract key numbers. A nil set holds none and takes
// no addition: it stands for a result whose reads are not recorded.
type readSet struct {
	words []uint64
}

func (s *readSet) add(id uint32) {
	w := int(id >> 6)
	if w >= len(s.words) {
		s.grow(w + 1)
	}
	s.words[w] |= 1 << (id & 63)
}

func (s *readSet) grow(n int) {
	if n <= cap(s.words) {
		s.words = s.words[:n]
		return
	}
	words := make([]uint64, n, max(n, 2*cap(s.words)))
	copy(words, s.words)
	s.words = words
}

func (s *readSet) has(id uint32) bool {
	if s == nil {
		return false
	}
	w := int(id >> 6)
	return w < len(s.words) && s.words[w]&(1<<(id&63)) != 0
}

// addAll adds what from holds.
func (s *readSet) addAll(from *readSet) {
	if from == nil || len(from.words) == 0 {
		return
	}
	if len(from.words) > len(s.words) {
		s.grow(len(from.words))
	}
	for i, w := range from.words {
		s.words[i] |= w
	}
}

// meets reports that s and other hold a common number.
func (s *readSet) meets(other *readSet) bool {
	if s == nil || other == nil {
		return false
	}
	n := min(len(s.words), len(other.words))
	for i := 0; i < n; i++ {
		if s.words[i]&other.words[i] != 0 {
			return true
		}
	}
	return false
}

func (s *readSet) empty() bool {
	if s == nil {
		return true
	}
	for _, w := range s.words {
		if w != 0 {
			return false
		}
	}
	return true
}

// each calls f with every number of s, in increasing order.
func (s *readSet) each(f func(id uint32)) {
	if s == nil {
		return
	}
	for i, w := range s.words {
		for w != 0 {
			b := bits.TrailingZeros64(w)
			f(uint32(i<<6 | b))
			w &= w - 1
		}
	}
}
