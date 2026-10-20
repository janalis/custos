package astquery

import (
	"bytes"
	"strings"

	"custos/internal/php/syntax"
)

type mentionsKey struct{ word string }

// Mentions reports whether the source of f spells word, ignoring ASCII
// case as PHP does for function, class and method names: a rule matching
// only a named builtin can leave a file that never spells it (calling it
// by a plain or imported name, or naming it in a string) before any
// costlier test. The answer is cached on the file per word.
func Mentions(f *syntax.File, word string) bool {
	return f.Memo(mentionsKey{strings.ToLower(word)}, func() any {
		return bytes.Contains(lowerSource(f), []byte(strings.ToLower(word)))
	}).(bool)
}

type lowerKey struct{}

func lowerSource(f *syntax.File) []byte {
	return f.Memo(lowerKey{}, func() any { return bytes.ToLower(f.Src) }).([]byte)
}
