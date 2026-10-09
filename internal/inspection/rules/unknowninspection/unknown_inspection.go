package unknowninspection

import (
	_ "embed"
	"strings"
	"sync"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/meta"
	"custos/internal/php/syntax"
)

// unknownInspection reports `@noinspection` tags naming inspections that
// neither custos nor PhpStorm know.
type unknownInspection struct{}

const unknownInspectionTag = "@noinspection"

//go:embed phpstorm_inspections.txt
var phpstormInspectionsTxt string

var phpstormInspections = sync.OnceValue(func() map[string]bool {
	m := map[string]bool{}
	for _, line := range strings.Split(phpstormInspectionsTxt, "\n") {
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		if line = strings.TrimSpace(line); line != "" {
			m[line] = true
		}
	}
	return m
})

func (unknownInspection) ID() string                           { return "UnknownInspection" }
func (unknownInspection) Kinds() []syntax.NodeKind             { return nil }
func (unknownInspection) Check(*analysis.Context, syntax.Node) {}
func (unknownInspection) CheckFile(ctx *analysis.Context) {
	src := ctx.Src
	for _, t := range ctx.File.Tokens {
		if t.Kind != syntax.TComment && t.Kind != syntax.TDocComment {
			continue
		}
		c := string(src[t.Start:t.End])
		if !strings.HasPrefix(c, "/*") { // D1: block comments only
			continue
		}
		for off := 0; ; {
			i := strings.Index(c[off:], unknownInspectionTag)
			if i < 0 {
				break
			}
			tagStart := off + i
			off = tagStart + len(unknownInspectionTag)
			value := c[off:] // D2
			if j := strings.IndexAny(value, "\r\n"); j >= 0 {
				value = value[:j]
			}
			if j := strings.Index(value, "*/"); j >= 0 {
				value = value[:j]
			}
			if unknown := unknownInspectionNames(value); len(unknown) > 0 {
				s := t.Start + uint32(tagStart)
				ctx.Report(syntax.Span{Start: s, End: s + uint32(len(unknownInspectionTag))},
					"Suppressed inspection not recognised: "+strings.Join(unknown, ", ")+".")
			}
		}
	}
}

// unknownInspectionNames returns the relevant but unknown inspection names
// of a tag value (D3–D5), in source order without duplicates.
func unknownInspectionNames(value string) []string {
	var out []string
	for i := 0; i < len(value); {
		if !unknownInspectionWordStart(value[i]) {
			i++
			continue
		}
		j := i + 1
		for j < len(value) && (unknownInspectionWordStart(value[j]) || value[j] >= '0' && value[j] <= '9') {
			j++
		}
		word := value[i:j]
		i = j
		if !strings.HasPrefix(word, "Php") && !strings.HasSuffix(word, "Inspection") && !strings.HasSuffix(word, "Inspector") {
			continue // D4
		}
		if knownInspection(word) || knownInspection(word+"Inspection") {
			continue
		}
		dup := false
		for _, o := range out {
			dup = dup || o == word
		}
		if !dup {
			out = append(out, word)
		}
	}
	return out
}

func knownInspection(name string) bool {
	if _, ok := meta.Lookup(name); ok {
		return true
	}
	return phpstormInspections()[name]
}

func unknownInspectionWordStart(c byte) bool {
	return c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}
