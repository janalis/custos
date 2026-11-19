// Package markup reads the expected findings that rule fixtures embed as
// inline tags around the highlighted code:
//
//	<warning descr="message">highlighted code</warning>
//
// Supported tags: error, warning, weak_warning (reported as info) and info.
// Tags may nest. A <caret> marker is accepted and removed. Both custos' own
// fixtures, the local EA checkout and the fixtures of a project's custom
// rules (custos test-rules) use this format.
package markup

import (
	"bytes"
	"fmt"
	"html"
	"sort"

	"github.com/janalis/custos/internal/diagnostic"
)

// Expectation is one highlighted range expected in the clean source.
type Expectation struct {
	Severity diagnostic.Severity
	Message  string
	Start    int // byte offset in the clean source, inclusive
	End      int // exclusive
}

func (e Expectation) String() string {
	return fmt.Sprintf("%s[%d:%d] %q", e.Severity, e.Start, e.End, e.Message)
}

var tagSeverity = map[string]diagnostic.Severity{
	"error":        diagnostic.SeverityError,
	"warning":      diagnostic.SeverityWarning,
	"weak_warning": diagnostic.SeverityInfo,
	"info":         diagnostic.SeverityInfo,
}

type openTag struct {
	name  string
	descr string
	start int
}

// Parse strips expectation tags from src and returns the clean source
// together with the expectations, sorted by position.
func Parse(src []byte) ([]byte, []Expectation, error) {
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
	Sort(out)
	return clean, out, nil
}

// Sort orders by start, end, severity, message.
func Sort(e []Expectation) {
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

// matchOpen recognizes `<name>` and `<name descr="...">` (attributes other
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

func key(e Expectation, withMsg bool) string {
	k := fmt.Sprintf("%s|%d|%d", e.Severity, e.Start, e.End)
	if withMsg {
		k += "|" + e.Message
	}
	return k
}

// Diff returns the expectations of want that got lacks and those of got
// that want lacks, counting duplicates; messages are compared when withMsg
// is set.
func Diff(want, got []Expectation, withMsg bool) (missing, unexpected []Expectation) {
	count := map[string]int{}
	for _, g := range got {
		count[key(g, withMsg)]++
	}
	for _, w := range want {
		k := key(w, withMsg)
		if count[k] > 0 {
			count[k]--
			continue
		}
		missing = append(missing, w)
	}
	seen := map[string]int{}
	for _, w := range want {
		seen[key(w, withMsg)]++
	}
	for _, g := range got {
		k := key(g, withMsg)
		if seen[k] > 0 {
			seen[k]--
			continue
		}
		unexpected = append(unexpected, g)
	}
	return missing, unexpected
}

// NormalizeWS collapses whitespace runs and trims, so fix comparisons
// ignore formatting differences.
func NormalizeWS(b []byte) []byte {
	return bytes.Join(bytes.Fields(b), []byte{' '})
}
