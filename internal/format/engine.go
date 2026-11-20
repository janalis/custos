package format

import (
	"strings"

	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
)

// engine formats one parsed file. Positions k index e.sig (significant
// tokens and string atoms, ending with EOF); gap k is the trivia between
// unit k-1 and unit k.
type engine struct {
	src  []byte
	f    *syntax.File
	s    *Settings
	php  phpversion.Version
	toks []syntax.Token

	// Per token (index into toks).
	role      []role
	br        []bracket
	start     []bool // begins a statement, member or list item
	ident     []bool // inside a name or identifier
	typeName  []bool
	constName []bool

	regions  []region // lines indented one more level
	bodies   []region // statement bodies without braces, indented when on their own line
	kept     []region // lines keeping their indentation
	destruct []bool   // per token: opening bracket of a destructuring list

	// Per unit k.
	sig      []int    // token index of the unit's first token
	end      []uint32 // end offset of the unit (atoms span several tokens)
	match    []int    // matching bracket unit, or -1
	text     []string // output text of the unit
	textRule []string
	unitOf   []int // per token: unit it starts, or -1
	gaps     []gap
	want     []int // indentation level of unit k if it begins a line; keepLevel when kept
	align    []int // columns of alignment after the indentation of unit k
	lineOf   []int // indentation level of the line holding unit k
	deleted  []bool
	outPos   []int // output offset of unit k after the last emit

	joinChecked map[[2]string]bool

	template bool   // mixed PHP and markup: layout is kept
	merged   bool   // a rewrite joined several units into one
	final    int    // gap after the last code unit, ending the file; -1 when kept
	eol      string // line break written
	ignored  []region
}

// gap is the decision for the trivia before one unit.
type gap struct {
	start, end uint32
	pieces     []piece // whitespace and comments, in order
	comment    bool
	origNL     int // line breaks in the leading whitespace (all of it without comments)
	lead       int // piece holding the leading whitespace
	tailNL     int // line breaks after the last comment
	nl         int // wanted line breaks in the leading whitespace; -1 keeps
	sp         int // wanted spaces when on one line; -1 keeps
	rule       string
	frozen     bool // emitted as written
	drop       bool // emitted empty (the unit before or after it is removed)
	dropLead   bool // the whitespace before its first comment is emitted empty
	toComment  bool // a comment of a later gap starts the line it opens
	into       int  // with dropLead, the gap whose line breaks it takes
	tailSet    int  // wanted line breaks after the last comment; -1 keeps
	tailRule   string
}

type piece struct {
	comment bool
	start   uint32
	end     uint32
}

// lineBreaks counts line breaks (\r\n, \n, \r) in b.
func lineBreaks(b []byte) int {
	n := 0
	for i := 0; i < len(b); i++ {
		switch b[i] {
		case '\n':
			n++
		case '\r':
			n++
			if i+1 < len(b) && b[i+1] == '\n' {
				i++
			}
		}
	}
	return n
}

// units groups the token stream into significant units. A string with
// interpolation, a heredoc or nowdoc and a shell command are one unit.
func (e *engine) units() {
	toks := e.toks
	for i := 0; i < len(toks); i++ {
		t := toks[i]
		if t.Kind.IsTrivia() {
			continue
		}
		e.sig = append(e.sig, i)
		j := i
		switch t.Kind {
		case syntax.TDoubleQuote, syntax.TStartHeredoc, syntax.TBacktick:
			j = e.stringEnd(i)
		}
		e.end = append(e.end, toks[j].End)
		i = j
	}
}

// stringEnd returns the index of the token closing the string opened at i.
func (e *engine) stringEnd(i int) int {
	closer := map[syntax.TokenKind]syntax.TokenKind{
		syntax.TDoubleQuote:  syntax.TDoubleQuote,
		syntax.TStartHeredoc: syntax.TEndHeredoc,
		syntax.TBacktick:     syntax.TBacktick,
	}
	stack := []syntax.TokenKind{closer[e.toks[i].Kind]}
	depth := []int{0}
	for j := i + 1; j < len(e.toks); j++ {
		k := e.toks[j].Kind
		top := len(stack) - 1
		switch {
		case depth[top] == 0 && k == stack[top]:
			stack, depth = stack[:top], depth[:top]
			if len(stack) == 0 {
				return j
			}
		case k == syntax.TCurlyOpen || k == syntax.TDollarOpenCurlyBraces:
			depth[top]++
		case depth[top] > 0 && k == syntax.TLBrace:
			depth[top]++
		case depth[top] > 0 && k == syntax.TRBrace:
			depth[top]--
		case depth[top] > 0 && (k == syntax.TDoubleQuote || k == syntax.TStartHeredoc || k == syntax.TBacktick):
			stack = append(stack, closer[k])
			depth = append(depth, 0)
		case k == syntax.TEOF:
			return j - 1
		}
	}
	return len(e.toks) - 1
}

// kind returns the token kind of unit k.
func (e *engine) kind(k int) syntax.TokenKind { return e.toks[e.sig[k]].Kind }

// orig returns the source text of unit k.
func (e *engine) orig(k int) string {
	return string(e.src[e.toks[e.sig[k]].Start:e.end[k]])
}

func opener(k syntax.TokenKind) bool {
	return k == syntax.TLParen || k == syntax.TLBracket || k == syntax.TLBrace || k == syntax.TAttribute
}

func closer(k syntax.TokenKind) bool {
	return k == syntax.TRParen || k == syntax.TRBracket || k == syntax.TRBrace
}

// pair matches brackets. A file that parses without error balances them.
func (e *engine) pair() {
	e.match = make([]int, len(e.sig))
	var stack []int
	for k := range e.sig {
		e.match[k] = -1
		kd := e.kind(k)
		switch {
		case opener(kd):
			stack = append(stack, k)
		case closer(kd) && len(stack) > 0:
			o := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			e.match[o], e.match[k] = k, o
		}
	}
}

// buildGaps records the trivia before every unit.
func (e *engine) buildGaps() {
	e.gaps = make([]gap, len(e.sig))
	prevEnd := uint32(0)
	for k := range e.sig {
		g := &e.gaps[k]
		g.start, g.end = prevEnd, e.toks[e.sig[k]].Start
		g.nl, g.sp, g.tailSet = -1, -1, -1
		for i := e.index(g.start); i < len(e.toks) && e.toks[i].Start < g.end; i++ {
			t := e.toks[i]
			g.pieces = append(g.pieces, piece{comment: t.Kind != syntax.TWhitespace, start: t.Start, end: t.End})
			if t.Kind != syntax.TWhitespace {
				g.comment = true
			}
		}
		// The leading whitespace is the first line break: after the comments
		// that trail the previous unit on its line.
		for i, p := range g.pieces {
			if p.comment {
				continue
			}
			if n := lineBreaks(e.src[p.start:p.end]); n > 0 || i == len(g.pieces)-1 {
				g.lead, g.origNL = i, n
				break
			}
		}
		if last := len(g.pieces) - 1; g.comment && last >= 0 && !g.pieces[last].comment {
			g.tailNL = lineBreaks(e.src[g.pieces[last].start:g.pieces[last].end])
		}
		prevEnd = e.end[k]
	}
}

// startsLine reports whether unit k begins a line in the output.
func (e *engine) startsLine(k int) bool {
	if k == 0 {
		return true
	}
	g := &e.gaps[k]
	switch {
	case g.frozen:
		return e.origStartsLine(k)
	case g.comment:
		last := g.pieces[len(g.pieces)-1]
		if last.comment {
			return e.commentEndsLine(last)
		}
		return e.tailBreaks(g) > 0
	}
	if e.leadBreaks(g) > 0 {
		return true
	}
	return endsLine(e.text[k-1])
}

// origStartsLine reports whether unit k begins a line in the source.
func (e *engine) origStartsLine(k int) bool {
	off := e.toks[e.sig[k]].Start
	for i := int(off) - 1; i >= 0; i-- {
		switch e.src[i] {
		case '\n', '\r':
			return true
		case ' ', '\t':
			continue
		}
		return false
	}
	return true
}

func (e *engine) commentEndsLine(p piece) bool {
	c := e.src[p.start:p.end]
	return len(c) > 0 && (c[0] == '#' || (len(c) > 1 && c[1] == '/'))
}

// leadBreaks is the number of line breaks gap g's leading whitespace gets.
func (e *engine) leadBreaks(g *gap) int {
	n := g.origNL
	if g.nl >= 0 {
		n = g.nl
	}
	return e.cap(n)
}

func (e *engine) tailBreaks(g *gap) int {
	n := g.tailNL
	if g.tailSet >= 0 {
		n = g.tailSet
	}
	return e.cap(n)
}

// cap limits n line breaks to the allowed blank lines.
func (e *engine) cap(n int) int {
	if e.s.MaxBlankLines > 0 && n > e.s.MaxBlankLines+1 {
		return e.s.MaxBlankLines + 1
	}
	return n
}

// setNL sets the line breaks before unit k (leading whitespace).
func (e *engine) setNL(k, n int, rule string) {
	k = e.written(k)
	if k <= 0 || k >= len(e.gaps) || e.gaps[k].frozen {
		return
	}
	g := &e.gaps[k]
	if g.comment && n == 0 {
		return
	}
	g.nl, g.rule = n, rule
	if !g.comment {
		g.sp = -1
		if n == 0 {
			g.sp = 1
		}
	}
}

// setTail sets the line breaks after the last comment of gap k, or of
// the whole gap without comment.
func (e *engine) setTail(k, n int, rule string) {
	if k <= 0 || k >= len(e.gaps) || e.gaps[k].frozen {
		return
	}
	g := &e.gaps[k]
	if !g.comment {
		e.setNL(k, n, rule)
		return
	}
	if last := g.pieces[len(g.pieces)-1]; last.comment {
		return
	}
	g.tailSet, g.tailRule = n, rule
}

// setSpace sets the spaces of a gap on one line.
func (e *engine) setSpace(k, n int, rule string) {
	g := &e.gaps[k]
	if g.frozen || g.comment {
		return
	}
	g.nl, g.sp, g.rule = 0, n, rule
}

// endsLine reports whether text ends with a line break.
func endsLine(text string) bool {
	return strings.HasSuffix(text, "\n") || strings.HasSuffix(text, "\r")
}

// inString reports whether off lies inside a string unit (an interpolated
// string, a heredoc or a shell command), past its first byte: code there is
// part of the string and is never rewritten.
func (e *engine) inString(off uint32) bool {
	lo, hi := 0, len(e.sig)
	for lo < hi {
		m := int(uint(lo+hi) >> 1)
		if e.toks[e.sig[m]].Start <= off {
			lo = m + 1
		} else {
			hi = m
		}
	}
	k := lo - 1
	return k >= 0 && off > e.toks[e.sig[k]].Start && off < e.end[k]
}

// written returns the gap written in place of gap k: k itself, or, when k
// is dropped after removed units, the gap before those units.
func (e *engine) written(k int) int {
	if k <= 0 || k >= len(e.gaps) {
		return k
	}
	if e.gaps[k].dropLead {
		return e.gaps[k].into
	}
	if !e.gaps[k].drop {
		return k
	}
	j := k
	for j > 1 && e.gaps[j].drop && e.deleted[j-1] {
		j--
	}
	if !e.gaps[j].drop {
		return j
	}
	for j = k; j < len(e.gaps)-1 && e.gaps[j].drop && e.deleted[j]; j++ {
	}
	return j
}

// verbatim reports whether gap k is written as in the source.
func (e *engine) verbatim(k int) bool {
	g := &e.gaps[k]
	return g.frozen || e.inRegions(e.ignored, g.start) || (g.end > 0 && e.inRegions(e.ignored, g.end-1))
}
