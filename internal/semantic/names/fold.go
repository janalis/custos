package names

import (
	"strings"
	"sync/atomic"

	"custos/internal/semantic/readmap"
)

// maxFolded bounds the names Fold remembers, and maxFoldedBytes their length,
// so a long-running server never grows it without limit; past them, names
// are folded on each call.
const (
	maxFolded      = 1 << 16
	maxFoldedBytes = 256
)

var (
	folded      readmap.Map[string, string] // name -> lower-cased name
	foldedCount atomic.Int64
)

// Fold returns name lower-cased, as PHP compares class, function and method
// names. Lookups repeat the same few thousand names, so their folded form is
// remembered instead of allocated again on every call.
func Fold(name string) string {
	if lowerASCII(name) {
		// Most names the index folds are its own keys: nothing to do.
		return name
	}
	if len(name) > maxFoldedBytes {
		return strings.ToLower(name)
	}
	if v, ok := folded.Load(name); ok {
		return v
	}
	lower := strings.ToLower(name)
	if lower != name && foldedCount.Load() < maxFolded {
		// Clone the key: name may share the memory of a whole source file.
		if folded.Store(strings.Clone(name), lower) {
			foldedCount.Add(1)
		}
	}
	return lower
}

// lowerASCII reports a name of ASCII bytes without an upper-case letter,
// which lower-casing leaves as it is. It tests eight bytes at a time: names
// are mostly long namespaced keys.
func lowerASCII(name string) bool {
	const ones = 0x0101010101010101
	i := 0
	for ; i+8 <= len(name); i += 8 {
		w := uint64(name[i]) | uint64(name[i+1])<<8 | uint64(name[i+2])<<16 | uint64(name[i+3])<<24 |
			uint64(name[i+4])<<32 | uint64(name[i+5])<<40 | uint64(name[i+6])<<48 | uint64(name[i+7])<<56
		// A byte has its high bit set when it is not ASCII; adding 0x3f to
		// an ASCII byte sets it from 'A' on, adding 0x25 from past 'Z' on,
		// without a carry into the next byte.
		if (w|(w+0x3f*ones)&^(w+0x25*ones))&(0x80*ones) != 0 {
			return false
		}
	}
	for ; i < len(name); i++ {
		if c := name[i]; c >= 0x80 || 'A' <= c && c <= 'Z' {
			return false
		}
	}
	return true
}
