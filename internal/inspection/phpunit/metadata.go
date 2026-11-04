package phpunit

import (
	"strings"

	"custos/internal/php/syntax"
)

// Metadata is a normalized PHPUnit declaration annotation or attribute.
type Metadata struct {
	Name       string // with '@'
	Start, End uint32 // absolute span of the tag Name
	Value      string // first word after the tag Name ("" when none)
	Attribute  bool   // originated from a resolved attribute
	Annotation bool   // starts a docblock line (D3)
}

func metadataTagChar(c byte) bool {
	return c == '_' || c == '-' || c == '\\' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
}

// Tags lists the tags of a doc comment token.
func Tags(f *syntax.File, tok syntax.Token) []Metadata {
	text := string(f.Src[tok.Start:tok.End])
	var out []Metadata
	for i := 0; i < len(text); i++ {
		if text[i] != '@' || (i > 0 && metadataTagChar(text[i-1])) {
			continue
		}
		j := i + 1
		for j < len(text) && metadataTagChar(text[j]) {
			j++
		}
		if j == i+1 {
			continue
		}
		t := Metadata{Name: text[i:j], Start: tok.Start + uint32(i), End: tok.Start + uint32(j)}
		// D3: Annotation position
		k := i
		for k > 0 && (text[k-1] == ' ' || text[k-1] == '\t' || text[k-1] == '\r' || text[k-1] == '\n') {
			k--
		}
		if k == 3 && strings.HasPrefix(text, "/**") {
			t.Annotation = true
		} else if k > 0 && text[k-1] == '*' {
			b := k - 1
			for b > 0 && (text[b-1] == ' ' || text[b-1] == '\t') {
				b--
			}
			t.Annotation = b > 0 && text[b-1] == '\n'
		}
		// Value: first word on the same line
		v := j
		for v < len(text) && (text[v] == ' ' || text[v] == '\t') {
			v++
		}
		w := v
		for w < len(text) && text[w] != ' ' && text[w] != '\t' && text[w] != '\n' && text[w] != '\r' &&
			!strings.HasPrefix(text[w:], "*/") {
			w++
		}
		t.Value = text[v:w]
		if t.Name == "@depends" && (t.Value == "clone" || t.Value == "shallowClone") {
			v = w
			for v < len(text) && (text[v] == ' ' || text[v] == '\t') {
				v++
			}
			w = v
			for w < len(text) && text[w] != ' ' && text[w] != '\t' && text[w] != '\n' && text[w] != '\r' && !strings.HasPrefix(text[w:], "*/") {
				w++
			}
			t.Value = text[v:w]
		}
		out = append(out, t)
		i = j - 1
	}
	return out
}
