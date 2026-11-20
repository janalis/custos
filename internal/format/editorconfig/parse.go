package editorconfig

import (
	"bytes"
	"strings"
)

// File is one parsed `.editorconfig` file.
type File struct {
	// Root is true when the preamble sets root to true (case-insensitive).
	Root bool
	// Sections are the file's sections in source order.
	Sections []Section
}

// Section is one `[glob]` section and the pairs written under it.
type Section struct {
	// Glob is the section name as written between the brackets.
	Glob string
	// Pairs are the section's key-value pairs in source order.
	Pairs []Pair
}

// Pair is one key-value line. Key is lowercased; Value is trimmed and kept as
// written.
type Pair struct {
	Key   string
	Value string
}

// Parse parses the content of one `.editorconfig` file. Parsing never fails:
// blank lines, comment lines (starting with `#` or `;`) and invalid lines are
// skipped, a `#` or `;` after the start of a line is ordinary text, and pairs
// in the preamble other than root are ignored.
func Parse(content []byte) *File {
	content = bytes.TrimPrefix(content, []byte("\uFEFF"))
	f := &File{}
	for line := range strings.SplitSeq(string(content), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case line == "", line[0] == '#', line[0] == ';':
		case line[0] == '[' && len(line) > 1 && line[len(line)-1] == ']':
			f.Sections = append(f.Sections, Section{Glob: line[1 : len(line)-1]})
		default:
			key, value, ok := strings.Cut(line, "=")
			key = strings.ToLower(strings.TrimSpace(key))
			if !ok || key == "" {
				continue
			}
			value = strings.TrimSpace(value)
			if n := len(f.Sections); n > 0 {
				f.Sections[n-1].Pairs = append(f.Sections[n-1].Pairs, Pair{Key: key, Value: value})
			} else if key == "root" {
				f.Root = strings.EqualFold(value, "true")
			}
		}
	}
	return f
}
