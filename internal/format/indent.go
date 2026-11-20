package format

import (
	"strings"

	"github.com/janalis/custos/internal/php/syntax"
)

// keepLevel marks a line that keeps its indentation as written.
const keepLevel = -1

type openBracket struct {
	k         int
	line      int  // level its content is indented from
	item      int  // level of the line where the current statement or item began
	itemOrig  int  // the same line's level as written
	itemCols  int  // the same line's indentation width as written
	itemAlign int  // alignment columns of that line in the output
	hanging   bool // content follows the bracket on its line
}

// indentation computes the level of every unit as if it began a line;
// e.want[k] is used for units that do.
func (e *engine) indentation() {
	n := len(e.sig)
	e.want = make([]int, n)
	e.align = make([]int, n)
	e.lineOf = make([]int, n)
	stack := []openBracket{{k: -1, line: -1}}
	cur, curAlign := 0, 0
	for k := range e.sig {
		lvl := e.levelFor(k, stack)
		if e.s.ContinuationIndent == "keep" && lvl != keepLevel && e.kind(k) != syntax.TEOF && e.origStartsLine(k) && e.isContinuation(k, stack) && stack[len(stack)-1].itemOrig >= 0 {
			f := stack[len(stack)-1]
			u := max(e.unit(), 1)
			cols := max(f.item, 0)*u + f.itemAlign + e.width(e.origIndent(k)) - f.itemCols
			lvl = max(f.item, 0)
			if cols > lvl*u {
				lvl, e.align[k] = cols/u, cols%u
			}
		}
		e.want[k] = lvl
		if e.commentStartsLine(k) && lvl != keepLevel {
			cur, curAlign = lvl, 0 // the comment opening the line is indented like k
		}
		if e.startsLine(k) {
			curAlign = e.align[k]
			if lvl == keepLevel {
				cur = e.origLevel(k)
			} else {
				cur = lvl
			}
		}
		e.lineOf[k] = cur
		if top := &stack[len(stack)-1]; e.start[e.sig[k]] || e.newItem(k) {
			top.item, top.itemOrig, top.itemCols, top.itemAlign = cur, e.measuredLevel(k), e.origCols(k), curAlign
		}
		switch kd := e.kind(k); {
		case opener(kd):
			line := cur
			if kd == syntax.TLBrace && e.br[e.sig[k]] != brNone && e.br[e.sig[k]] != brExprBrace {
				// A body is indented from the line its construct began on.
				line = e.lineOf[e.owner(k)]
			}
			hanging := e.match[k] > k+1 && !e.startsLine(k+1) && kd != syntax.TLBrace
			stack = append(stack, openBracket{k: k, line: line, item: cur, itemOrig: e.measuredLevel(k), itemCols: e.origCols(k), itemAlign: curAlign, hanging: hanging})
		case closer(kd):
			if len(stack) > 1 && stack[len(stack)-1].k == e.match[k] {
				stack = stack[:len(stack)-1]
			}
		}
		if strings.ContainsAny(e.text[k], "\n\r") && (e.kind(k) == syntax.TStartHeredoc) {
			cur = e.lastLineLevel(k)
		}
	}
}

// levelFor returns the indentation level of unit k at the start of a line.
func (e *engine) levelFor(k int, stack []openBracket) int {
	off := e.toks[e.sig[k]].Start
	if e.template || e.s.IndentStyle == "keep" || e.inRegions(e.kept, off) || e.inRegions(e.ignored, off) {
		return keepLevel
	}
	if e.kind(k) == syntax.TEOF {
		return 0
	}
	if k > 0 && e.verbatim(k) {
		return keepLevel // its gap is written as is
	}
	frame := &stack[len(stack)-1]
	var top *openBracket
	if frame.k >= 0 {
		top = frame
	}
	if top != nil && closer(e.kind(k)) && e.match[k] == top.k {
		return top.line
	}
	if top != nil && top.hanging {
		return e.hangLevel(stack)
	}
	base, topOff, topBr := 0, uint32(0), brNone
	if top != nil {
		base = top.line + 1
		topOff = e.toks[e.sig[top.k]].Start
		topBr = e.br[e.sig[top.k]]
		if e.kind(top.k) == syntax.TAttribute {
			topBr = brAttr
		}
	}
	extra := 0
	for _, r := range e.regions {
		if r.start <= off && off < r.end && (top == nil || r.start > topOff) {
			extra++
		}
	}
	for _, r := range e.bodies {
		if r.start <= off && off < r.end && (top == nil || r.start > topOff) && e.bodyOnOwnLine(r) {
			extra++
		}
	}
	if e.continuation(k, top, topBr) == 1 {
		// Continuation lines are indented once more than the line where the
		// statement or item began; deeper continuations written for nested
		// operands keep their extra depth.
		depth := 1
		chain := e.kind(k) == syntax.TObjectOperator || e.kind(k) == syntax.TNullsafeObjectOperator
		if e.origStartsLine(k) && frame.itemOrig >= 0 && !chain {
			depth = max(1, e.origLevel(k)-frame.itemOrig)
		}
		return max(base+extra, frame.item+depth)
	}
	return base + extra
}

// owner returns the unit beginning the construct whose body the brace k
// opens: the statement, member, closure, match or anonymous class.
func (e *engine) owner(k int) int {
	j := k - 1
	for j > 0 {
		t := e.sig[j]
		kd := e.kind(j)
		switch {
		case closer(kd) && e.match[j] >= 0:
			j = e.match[j] - 1
			continue
		case kd == syntax.TSemicolon || kd == syntax.TLBrace || kd == syntax.TRBrace || opener(kd) || kd == syntax.TComma && !e.inSignature(j):
			return j + 1
		case e.start[t] || e.role[t] == roleClosureKeyword || e.role[t] == roleArrowKeyword || e.role[t] == roleNewKeyword:
			return j
		}
		switch kd {
		case syntax.TMatch, syntax.TElse, syntax.TElseif, syntax.TCatch, syntax.TFinally, syntax.TDo, syntax.TTry:
			return j
		}
		j--
	}
	return max(j, 0)
}

// inSignature reports whether the comma j separates the names of an
// extends or implements list.
func (e *engine) inSignature(j int) bool {
	for i := j - 1; ; i-- { // stops at the opening tag at the latest
		switch e.kind(i) {
		case syntax.TImplements, syntax.TExtends:
			return true
		case syntax.TString, syntax.TNameQualified, syntax.TNameFullyQualified, syntax.TNameRelative, syntax.TComma:
			continue
		}
		return false
	}
}

// hangLevel is the level of a line inside a bracket whose content starts
// on the bracket's own line: the level a continuation line gets in the
// enclosing context.
func (e *engine) hangLevel(stack []openBracket) int {
	for len(stack) > 1 && stack[len(stack)-1].hanging {
		stack = stack[:len(stack)-1]
	}
	f := stack[len(stack)-1]
	if f.k < 0 {
		return f.item + 1
	}
	b := e.br[e.sig[f.k]]
	if e.kind(f.k) == syntax.TAttribute {
		b = brAttr
	}
	if b.block() || b.list() || b == brHooks {
		return max(f.line+1, f.item+1)
	}
	return f.line + 1
}

// newItem reports whether unit k begins a list item: it follows the
// opening bracket or a comma of its list.
func (e *engine) newItem(k int) bool {
	if k == 0 {
		return true
	}
	p := k - 1
	switch e.kind(p) {
	case syntax.TComma, syntax.TSemicolon, syntax.TOpenTag:
		return true
	case syntax.TLBrace:
		return e.br[e.sig[p]] != brNone && e.br[e.sig[p]] != brExprBrace
	case syntax.TRBrace:
		o := e.match[p]
		return o >= 0 && e.br[e.sig[o]].block()
	}
	return opener(e.kind(p)) && !closer(e.kind(k))
}

// continuation is 1 when unit k continues the statement or list item
// begun on an earlier line.
func (e *engine) continuation(k int, top *openBracket, topBr bracket) int {
	t := e.sig[k]
	if e.start[t] || e.kind(k) == syntax.TLBrace || closer(e.kind(k)) || k == 0 {
		return 0
	}
	if topBr.list() && e.newItem(k) {
		return 0
	}
	switch e.kind(k - 1) {
	case syntax.TOpenTag, syntax.TOpenTagWithEcho, syntax.TCloseTag, syntax.TInlineHTML:
		return 0
	}
	if top == nil || topBr.block() || topBr == brHooks || topBr.list() {
		return 1
	}
	return 0
}

func (e *engine) inRegions(rs []region, off uint32) bool {
	for _, r := range rs {
		if r.start <= off && off < r.end {
			return true
		}
	}
	return false
}

// bodyOnOwnLine reports whether the statement body r starts a line.
func (e *engine) bodyOnOwnLine(r region) bool {
	k := e.unitAt(r.start)
	return k >= 0 && e.startsLine(k)
}

// origIndent returns the indentation written before unit k.
func (e *engine) origIndent(k int) string {
	off := int(e.toks[e.sig[k]].Start)
	i := off
	for i > 0 && (e.src[i-1] == ' ' || e.src[i-1] == '\t') {
		i--
	}
	if i > 0 && e.src[i-1] != '\n' && e.src[i-1] != '\r' {
		return ""
	}
	return string(e.src[i:off])
}

// width returns the display width of indentation ind.
func (e *engine) width(ind string) int {
	w := 0
	for _, c := range ind {
		if c == '\t' {
			w += max(e.s.TabWidth, 1) - w%max(e.s.TabWidth, 1)
		} else {
			w++
		}
	}
	return w
}

// isContinuation reports whether unit k, starting a line, continues the
// statement or item begun on an earlier line of the innermost bracket.
func (e *engine) isContinuation(k int, stack []openBracket) bool {
	f := stack[len(stack)-1]
	var top *openBracket
	b := brNone
	if f.k >= 0 {
		top = &f
		b = e.br[e.sig[f.k]]
		if f.hanging {
			return !closer(e.kind(k)) || e.match[k] != f.k
		}
	}
	return e.continuation(k, top, b) == 1
}

// origCols is the indentation width of the line holding unit k as written.
func (e *engine) origCols(k int) int {
	return e.origLevel(k)*max(e.unit(), 1) + e.origRest(k)
}

// origRest is the indentation width of k's line beyond whole levels.
func (e *engine) origRest(k int) int {
	return e.lineWidth(k) % max(e.unit(), 1)
}

// lineWidth is the indentation width of the line holding unit k.
func (e *engine) lineWidth(k int) int {
	return e.width(e.lineLead(k))
}

// origLevel estimates the level of the line holding unit k as written.
func (e *engine) origLevel(k int) int {
	return e.lineWidth(k) / max(e.unit(), 1)
}

// lastLineLevel estimates the level of the last line of a heredoc unit.
func (e *engine) lastLineLevel(k int) int {
	t := e.text[k]
	i := strings.LastIndexAny(t, "\n\r")
	rest := t[i+1:]
	ind := rest[:len(rest)-len(strings.TrimLeft(rest, " \t"))]
	return e.width(ind) / max(e.unit(), 1)
}

// unit is the width of one indentation level.
func (e *engine) unit() int {
	if e.s.IndentStyle == "tab" {
		return max(e.s.TabWidth, 1)
	}
	return e.s.IndentSize
}

// indentString renders level lvl.
func (e *engine) indentString(lvl int) string {
	if lvl <= 0 {
		return ""
	}
	if e.s.IndentStyle == "tab" {
		return strings.Repeat("\t", lvl)
	}
	return strings.Repeat(" ", lvl*e.s.IndentSize)
}

// lineIndent returns the indentation written before unit k when it begins
// a line.
func (e *engine) lineIndent(k int) string {
	if e.want[k] == keepLevel {
		return e.origIndent(k)
	}
	return e.indentString(e.want[k]) + strings.Repeat(" ", e.align[k])
}

// commentIndent returns the indentation of a comment line before unit k.
func (e *engine) commentIndent(k int) string {
	if e.want[k] == keepLevel {
		return ""
	}
	lvl := e.want[k]
	if closer(e.kind(k)) && e.match[k] >= 0 {
		lvl++
	}
	return e.indentString(lvl)
}

// lineLead returns the indentation of the source line holding unit k.
func (e *engine) lineLead(k int) string {
	off := int(e.toks[e.sig[k]].Start)
	i := off
	for i > 0 && e.src[i-1] != '\n' && e.src[i-1] != '\r' {
		i--
	}
	j := i
	for j < off && (e.src[j] == ' ' || e.src[j] == '\t') {
		j++
	}
	return string(e.src[i:j])
}

// measuredLevel is the level of unit k's source line, or -1 when that line
// begins inside a comment or a string, whose text formatting never moves.
func (e *engine) measuredLevel(k int) int {
	if e.startsLine(k) && !e.origStartsLine(k) || e.commentStartsLine(k) {
		return -1 // moved to a line of its own, or after a comment: its source line measures nothing
	}
	off := int(e.toks[e.sig[k]].Start)
	ls := off
	for ls > 0 && e.src[ls-1] != '\n' && e.src[ls-1] != '\r' {
		ls--
	}
	if i := e.index(uint32(ls)); i > 0 {
		if t := e.toks[i-1]; t.End > uint32(ls) && t.Kind != syntax.TWhitespace {
			return -1
		}
	}
	if u := e.unitBefore(uint32(ls) + 1); u >= 0 && e.toks[e.sig[u]].Start < uint32(ls) && e.end[u] > uint32(ls) {
		return -1 // inside a string unit
	}
	return e.origLevel(k)
}

// commentStartsLine reports whether unit k follows, on the same line, a
// comment that begins a line.
func (e *engine) commentStartsLine(k int) bool {
	if k == 0 || !e.gaps[k].comment || e.gaps[k].frozen {
		return false
	}
	g := &e.gaps[k]
	n := len(g.pieces)
	if !g.pieces[n-1].comment {
		return false
	}
	// The run of comments before k, back to a line break.
	for i := n - 1; i >= 0; i-- {
		p := g.pieces[i]
		if p.comment {
			continue // a // comment always ends at a line break, whitespace
		}
		return lineBreaks(e.src[p.start:p.end]) > 0 || i == g.lead && e.leadBreaks(g) > 0 || i == 0 && endsLine(e.text[k-1])
	}
	return endsLine(e.text[k-1])
}
