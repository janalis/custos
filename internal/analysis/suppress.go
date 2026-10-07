package analysis

import (
	"sort"
	"strings"

	"custos/internal/meta"
	"custos/internal/syntax"
)

// Suppression comments (PhpStorm compatible):
//
//	/** @noinspection RuleIDInspection */  before a statement/declaration
//	// @noinspection RuleIDInspection     same, line comment
//	/** @noinspection ALL */
//	// @custos-ignore RuleID
//
// A suppression comment placed before the first statement of the file
// applies to the whole file.

type suppressions struct {
	f       *syntax.File
	fileIDs map[string]bool
	// cache: statement start -> suppressed IDs (nil when none)
	cache map[uint32]map[string]bool
}

func newSuppressions(f *syntax.File) *suppressions {
	s := &suppressions{f: f, cache: map[uint32]map[string]bool{}}
	if len(f.Stmts) > 0 {
		s.fileIDs = s.commentIDsBefore(f.Stmts[0].Span().Start)
		// An opening tag statement is not a node; also look before the first real statement.
	}
	return s
}

func (s *suppressions) suppressed(fd Finding) bool {
	m, _ := meta.Lookup(fd.Rule)
	match := func(ids map[string]bool) bool {
		if len(ids) == 0 {
			return false
		}
		return ids["ALL"] || ids[fd.Rule] || (m != nil && ids[m.LegacyID])
	}
	if match(s.fileIDs) {
		return true
	}
	for _, n := range s.enclosingStmts(fd.Span) {
		st := n.Span().Start
		ids, ok := s.cache[st]
		if !ok {
			ids = s.commentIDsBefore(st)
			s.cache[st] = ids
		}
		if match(ids) {
			return true
		}
	}
	return false
}

// enclosingStmts returns statement-like nodes containing span, outermost first.
func (s *suppressions) enclosingStmts(span syntax.Span) []syntax.Node {
	var out []syntax.Node
	var walk func(nodes func(func(syntax.Node)))
	walk = func(children func(func(syntax.Node))) {
		children(func(n syntax.Node) {
			if !n.Span().Contains(span) {
				return
			}
			if isSuppressible(n) {
				out = append(out, n)
			}
			walk(func(fn func(syntax.Node)) { syntax.Children(n, fn) })
		})
	}
	walk(func(fn func(syntax.Node)) {
		for _, st := range s.f.Stmts {
			fn(st)
		}
	})
	return out
}

func isSuppressible(n syntax.Node) bool {
	switch n.(type) {
	case syntax.Stmt, *syntax.Param, *syntax.PropertyHook:
		return true
	}
	return false
}

// commentIDsBefore collects suppressed IDs from the comments directly
// preceding offset (only trivia between them and the node).
func (s *suppressions) commentIDsBefore(offset uint32) map[string]bool {
	toks := s.f.Tokens
	i := sort.Search(len(toks), func(i int) bool { return toks[i].Start >= offset })
	var ids map[string]bool
	for j := i - 1; j >= 0; j-- {
		t := toks[j]
		if t.Kind == syntax.TWhitespace || t.Kind == syntax.TOpenTag {
			continue
		}
		if t.Kind != syntax.TComment && t.Kind != syntax.TDocComment {
			break
		}
		for _, id := range ParseSuppressionComment(string(s.f.Src[t.Start:t.End])) {
			if ids == nil {
				ids = map[string]bool{}
			}
			ids[id] = true
		}
	}
	return ids
}

// ParseSuppressionComment extracts suppressed rule names from a comment.
func ParseSuppressionComment(c string) []string {
	var out []string
	for _, tag := range []string{"@noinspection", "@custos-ignore"} {
		rest := c
		for {
			i := strings.Index(rest, tag)
			if i < 0 {
				break
			}
			rest = rest[i+len(tag):]
			line := rest
			if j := strings.IndexAny(line, "\n\r"); j >= 0 {
				line = line[:j]
			}
			line = strings.TrimSuffix(strings.TrimSpace(line), "*/")
			for _, f := range strings.FieldsFunc(line, func(r rune) bool { return r == ' ' || r == ',' || r == '\t' }) {
				if f == "*/" || f == "*" {
					continue
				}
				out = append(out, f)
			}
		}
	}
	return out
}
