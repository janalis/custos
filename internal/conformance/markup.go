// Package conformance checks rule implementations against annotated fixtures.
//
// Fixtures embed expectations as inline tags around the highlighted code:
//
//	<warning descr="message">highlighted code</warning>
//
// Supported tags: error, warning, weak_warning (reported as info) and info.
// Tags may nest. A <caret> marker is accepted and removed. Both custos' own
// fixtures and the local EA checkout use this format.
package conformance

import (
	"bytes"
	"fmt"
	"html"
	"sort"

	"custos/internal/meta"
)

// Expectation is one highlighted range expected in the clean source.
type Expectation struct {
	Severity meta.Severity
	Message  string
	Start    int // byte offset in the clean source, inclusive
	End      int // exclusive
}

func (e Expectation) String() string {
	return fmt.Sprintf("%s[%d:%d] %q", e.Severity, e.Start, e.End, e.Message)
}

var tagSeverity = map[string]meta.Severity{
	"error":        meta.SeverityError,
	"warning":      meta.SeverityWarning,
	"weak_warning": meta.SeverityInfo,
	"info":         meta.SeverityInfo,
}

type openTag struct {
	name  string
	descr string
	start int
}

// ParseMarkup strips expectation tags from src and returns the clean source
// together with the expectations, sorted by position.
func ParseMarkup(src []byte) ([]byte, []Expectation, error) {
	clean := make([]byte, 0, len(src))
	var stack []openTag
	var out []Expectation
	for i := 0; i < len(src); {
		if src[i] != '<' {
			j := bytes.IndexByte(src[i:], '<')
			if j < 0 {
				j = len(src) - i
			}
			clean = append(clean, src[i:i+j]...)
			i += j
			continue
		}
		if n, ok := matchLiteral(src[i:], "<caret>"); ok {
			i += n
			continue
		}
		if name, n, ok := matchClose(src[i:]); ok {
			if len(stack) == 0 || stack[len(stack)-1].name != name {
				return nil, nil, fmt.Errorf("offset %d: unbalanced </%s>", i, name)
			}
			top := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			out = append(out, Expectation{Severity: tagSeverity[name], Message: top.descr, Start: top.start, End: len(clean)})
			i += n
			continue
		}
		if name, descr, n, ok := matchOpen(src[i:]); ok {
			stack = append(stack, openTag{name: name, descr: descr, start: len(clean)})
			i += n
			continue
		}
		clean = append(clean, '<')
		i++
	}
	if len(stack) > 0 {
		return nil, nil, fmt.Errorf("unclosed <%s> at clean offset %d", stack[len(stack)-1].name, stack[len(stack)-1].start)
	}
	SortExpectations(out)
	return clean, out, nil
}

// SortExpectations orders by start, end, severity, message.
func SortExpectations(e []Expectation) {
	sort.Slice(e, func(i, j int) bool {
		a, b := e[i], e[j]
		if a.Start != b.Start {
			return a.Start < b.Start
		}
		if a.End != b.End {
			return a.End < b.End
		}
		if a.Severity != b.Severity {
			return a.Severity < b.Severity
		}
		return a.Message < b.Message
	})
}

func matchLiteral(b []byte, lit string) (int, bool) {
	if bytes.HasPrefix(b, []byte(lit)) {
		return len(lit), true
	}
	return 0, false
}

func matchClose(b []byte) (string, int, bool) {
	for name := range tagSeverity {
		if n, ok := matchLiteral(b, "</"+name+">"); ok {
			return name, n, true
		}
	}
	return "", 0, false
}

// matchOpen recognises `<name>` and `<name descr="...">` (attributes other
// than descr, such as textAttributesKey, are ignored).
func matchOpen(b []byte) (name, descr string, n int, ok bool) {
	for tag := range tagSeverity {
		p := "<" + tag
		if !bytes.HasPrefix(b, []byte(p)) || len(b) <= len(p) {
			continue
		}
		rest := b[len(p):]
		if rest[0] != '>' && rest[0] != ' ' {
			continue
		}
		i := 0
		for i < len(rest) && rest[i] != '>' {
			if rest[i] == '"' {
				j := bytes.IndexByte(rest[i+1:], '"')
				if j < 0 {
					return "", "", 0, false
				}
				i += j + 2
				continue
			}
			i++
		}
		if i == len(rest) {
			return "", "", 0, false
		}
		attrs := rest[:i]
		if k := bytes.Index(attrs, []byte(`descr="`)); k >= 0 {
			v := attrs[k+len(`descr="`):]
			if e := bytes.IndexByte(v, '"'); e >= 0 {
				descr = html.UnescapeString(string(v[:e]))
			}
		}
		return tag, descr, len(p) + i + 1, true
	}
	return "", "", 0, false
}
