package analysis

import (
	"bytes"
	"sort"
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/meta"
	"custos/internal/php/syntax"
)

// Suppression comments (PhpStorm compatible):
//
//	/** @noinspection RuleIDInspection */  before a statement/declaration
//	// @noinspection RuleIDInspection     same, line comment
//	/** @noinspection ALL */
//	// @custos-ignore RuleID
//
// A suppression comment placed before the first statement of the file
// applies to the whole file. Inside a statement, a comment right before a
// call argument or an array item applies to that argument or item only
// (a multi-line `new Row(…)` or array literal is one statement).

type suppressions struct {
	f       *syntax.File
	fileIDs map[string]bool
	// byTok caches commentIDsBefore by the index of the first token at or
	// after the offset (offsets sharing it share the preceding comments).
	byTok map[int]map[string]bool
	// sups are the suppressible nodes preceded by a suppression comment,
	// in tree preorder (start ascending, outer before inner); parent links
	// each to its nearest enclosing entry (-1 at the top).
	sups []suppressor
	memo map[supMemoKey]bool
}

type suppressor struct {
	span   syntax.Span
	ids    map[string]bool
	parent int
}

type supMemoKey struct {
	i    int
	rule string
}

func newSuppressions(f *syntax.File) *suppressions {
	s := &suppressions{f: f, byTok: map[int]map[string]bool{}}
	if !hasSuppressionComment(f) {
		return s // nothing can be suppressed: skip the tree walk
	}
	if len(f.Stmts) > 0 {
		s.fileIDs = s.commentIDsBefore(f.Stmts[0].Span().Start)
		// An opening tag statement is not a node; also look before the first real statement.
	}
	// One walk collects the suppressing nodes with their nesting, so each
	// finding is then resolved in O(log n + depth) instead of re-walking the
	// tree from the top (quadratic on large files with many findings).
	var stack []int // indexes into sups of the open (enclosing) entries
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if !isSuppressible(n) {
			return true
		}
		ids := s.commentIDsBefore(n.Span().Start)
		if len(ids) == 0 {
			return true
		}
		sp := n.Span()
		for len(stack) > 0 && !s.sups[stack[len(stack)-1]].span.Contains(sp) {
			stack = stack[:len(stack)-1]
		}
		parent := -1
		if len(stack) > 0 {
			parent = stack[len(stack)-1]
		}
		s.sups = append(s.sups, suppressor{span: sp, ids: ids, parent: parent})
		stack = append(stack, len(s.sups)-1)
		return true
	})
	return s
}

// hasSuppressionComment reports whether any comment of f carries a
// suppression tag.
func hasSuppressionComment(f *syntax.File) bool {
	for _, t := range f.Tokens {
		if t.Kind != syntax.TComment && t.Kind != syntax.TDocComment {
			continue
		}
		c := f.Src[t.Start:t.End]
		if bytes.Contains(c, []byte("@noinspection")) || bytes.Contains(c, []byte("@custos-ignore")) {
			return true
		}
	}
	return false
}

func (s *suppressions) suppressed(fd diagnostic.Finding) bool {
	if len(s.fileIDs) == 0 && len(s.sups) == 0 {
		return false
	}
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
	// The last entry starting at or before the finding; walk up to the
	// innermost entry containing it, whose enclosing chain all contains it.
	i := sort.Search(len(s.sups), func(i int) bool { return s.sups[i].span.Start > fd.Span.Start }) - 1
	for i >= 0 && !s.sups[i].span.Contains(fd.Span) {
		i = s.sups[i].parent
	}
	return s.chainMatches(i, fd.Rule, match)
}

// chainMatches reports whether entry i or one of its enclosing entries
// suppresses rule (memoized per entry and rule).
func (s *suppressions) chainMatches(i int, rule string, match func(map[string]bool) bool) bool {
	if i < 0 {
		return false
	}
	k := supMemoKey{i, rule}
	if v, ok := s.memo[k]; ok {
		return v
	}
	v := match(s.sups[i].ids) || s.chainMatches(s.sups[i].parent, rule, match)
	if s.memo == nil {
		s.memo = map[supMemoKey]bool{}
	}
	s.memo[k] = v
	return v
}

func isSuppressible(n syntax.Node) bool {
	switch n.(type) {
	case syntax.Stmt, *syntax.Param, *syntax.PropertyHook, *syntax.Arg, *syntax.ArrayItem:
		return true
	}
	return false
}

// commentIDsBefore collects suppressed IDs from the comments directly
// preceding offset (only trivia between them and the node).
func (s *suppressions) commentIDsBefore(offset uint32) map[string]bool {
	toks := s.f.Tokens
	i := sort.Search(len(toks), func(i int) bool { return toks[i].Start >= offset })
	if ids, ok := s.byTok[i]; ok {
		return ids
	}
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
	s.byTok[i] = ids
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
			// Further tags on this line are already part of its fields:
			// resume after the line (re-scanning it per tag was quadratic).
			rest = rest[len(line):]
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
