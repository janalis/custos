package names

import (
	"strings"
	"sync"
	"sync/atomic"
)

// maxFolded bounds the names Fold remembers, so a long-running server never
// grows it without limit; past it, names are folded on each call.
const maxFolded = 1 << 16

var (
	folded      sync.Map // name -> lower-cased name
	foldedCount atomic.Int64
)

// Fold returns name lower-cased, as PHP compares class, function and method
// names. Lookups repeat the same few thousand names, so their folded form is
// remembered instead of allocated again on every call.
func Fold(name string) string {
	if v, ok := folded.Load(name); ok {
		return v.(string)
	}
	lower := strings.ToLower(name)
	if lower != name && foldedCount.Load() < maxFolded {
		// Clone the key: name may share the memory of a whole source file.
		if _, loaded := folded.LoadOrStore(strings.Clone(name), lower); !loaded {
			foldedCount.Add(1)
		}
	}
	return lower
}
