// Package editorconfig resolves the EditorConfig properties of one file
// following https://spec.editorconfig.org/: it parses `.editorconfig` files,
// matches their section globs and merges their pairs by precedence. It reads
// files only through a caller-supplied function, so callers keep control of
// the IO.
package editorconfig

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
)

// Properties are the resolved key → value pairs for one file. Keys are
// lowercased; values of the known keys are lowercased, others are kept as
// written.
type Properties map[string]string

// known lists the keys whose values are case-insensitive.
var known = map[string]bool{
	"indent_style":             true,
	"indent_size":              true,
	"tab_width":                true,
	"end_of_line":              true,
	"charset":                  true,
	"trim_trailing_whitespace": true,
	"insert_final_newline":     true,
	"root":                     true,
	"max_line_length":          true,
}

// Resolve returns the properties that apply to path, an absolute OS or
// slash-separated file path. It reads the `.editorconfig` file of each
// directory from path's own up to the filesystem root, stopping after a file
// whose preamble sets root to true. read returns a file's content; an error
// matching fs.ErrNotExist skips the file, and any other error is returned.
// Pairs of nearer files override farther ones, and later sections override
// earlier ones within a file. A value of "unset" (any case) removes the key.
func Resolve(path string, read func(name string) ([]byte, error)) (Properties, error) {
	type found struct {
		file *File
		rel  string
	}
	var files []found
	p := filepath.Clean(filepath.FromSlash(path))
	dir, rel := filepath.Dir(p), filepath.Base(p)
	for {
		data, err := read(filepath.Join(dir, ".editorconfig"))
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
		if err == nil {
			f := Parse(data)
			files = append(files, found{f, rel})
			if f.Root {
				break
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		rel = filepath.Base(dir) + "/" + rel
		dir = parent
	}
	props := Properties{}
	for i := len(files) - 1; i >= 0; i-- {
		for _, sec := range files[i].file.Sections {
			if Match(sec.Glob, files[i].rel) {
				props.apply(sec.Pairs)
			}
		}
	}
	props.normalize()
	return props, nil
}

func (p Properties) apply(pairs []Pair) {
	for _, pair := range pairs {
		v := pair.Value
		if known[pair.Key] {
			v = strings.ToLower(v)
		}
		if strings.EqualFold(v, "unset") {
			delete(p, pair.Key)
			continue
		}
		p[pair.Key] = v
	}
}

// normalize derives the indentation defaults the specification implies.
func (p Properties) normalize() {
	if _, ok := p["indent_size"]; !ok && p["indent_style"] == "tab" {
		p["indent_size"] = "tab"
	}
	size := p["indent_size"]
	width, hasWidth := p["tab_width"]
	if _, err := strconv.Atoi(size); err == nil && !hasWidth {
		p["tab_width"] = size
	}
	if size == "tab" && hasWidth {
		p["indent_size"] = width
	}
}

// Settings is the formatter-relevant view of Properties; a nil field means the
// key is not set or holds an unsupported value.
type Settings struct {
	IndentStyle            *string // "tab" or "space"
	IndentSize             *int    // columns per indentation level
	TabWidth               *int    // columns per tab character
	EndOfLine              *string // "lf", "cr" or "crlf"
	InsertFinalNewline     *bool
	TrimTrailingWhitespace *bool
	MaxLineLength          *int // nil when unset or "off"
	Charset                *string
}

// Settings returns the typed view of p. An indent_size of "tab" takes the tab
// width when one is known.
func (p Properties) Settings() Settings {
	size := p["indent_size"]
	if size == "tab" {
		size = p["tab_width"]
	}
	return Settings{
		IndentStyle:            oneOf(p["indent_style"], "tab", "space"),
		IndentSize:             positive(size),
		TabWidth:               positive(p["tab_width"]),
		EndOfLine:              oneOf(p["end_of_line"], "lf", "cr", "crlf"),
		InsertFinalNewline:     boolean(p["insert_final_newline"]),
		TrimTrailingWhitespace: boolean(p["trim_trailing_whitespace"]),
		MaxLineLength:          positive(p["max_line_length"]),
		Charset:                oneOf(p["charset"], "latin1", "utf-8", "utf-8-bom", "utf-16be", "utf-16le"),
	}
}

func oneOf(v string, allowed ...string) *string {
	for _, a := range allowed {
		if v == a {
			return &a
		}
	}
	return nil
}

func positive(v string) *int {
	if n, err := strconv.Atoi(v); err == nil && n > 0 {
		return &n
	}
	return nil
}

func boolean(v string) *bool {
	if v != "true" && v != "false" {
		return nil
	}
	b := v == "true"
	return &b
}
