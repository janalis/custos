package format

import (
	"strings"

	"github.com/janalis/custos/internal/php/syntax"
)

// chunk is one piece of output replacing a span of the source.
type chunk struct {
	start, end uint32
	text       string
	rule       string
	// comment, when set, is rendered at write time, knowing whether it
	// starts its output line; k is the unit it precedes.
	comment *piece
	k       int
}

// emit renders the file and the edits it makes.
func (e *engine) emit() ([]byte, []Change) {
	var out strings.Builder
	out.Grow(len(e.src) + len(e.src)/16)
	var changes []Change
	add := func(c chunk) {
		if c.comment != nil {
			c = e.commentChunk(c.k, *c.comment, atLineStart(out.String()))
		}
		out.WriteString(c.text)
		if c.text == string(e.src[c.start:c.end]) {
			return
		}
		if n := len(changes); n > 0 && changes[n-1].End == c.start && changes[n-1].Rule == c.rule {
			changes[n-1].End = c.end
			changes[n-1].New += c.text
			return
		}
		changes = append(changes, Change{Rule: c.rule, Start: c.start, End: c.end, New: c.text})
	}
	e.outPos = make([]int, len(e.sig))
	for k := range e.sig {
		for _, c := range e.gapChunks(k) {
			add(c)
		}
		e.outPos[k] = out.Len()
		t := e.toks[e.sig[k]]
		add(chunk{start: t.Start, end: e.end[k], text: e.text[k], rule: e.textRule[k]})
	}
	return []byte(out.String()), changes
}

// gapChunks renders gap k.
func (e *engine) gapChunks(k int) []chunk {
	g := &e.gaps[k]
	orig := string(e.src[g.start:g.end])
	switch {
	case g.start == g.end && k == 0:
		return nil
	case e.verbatim(k):
		return []chunk{{start: g.start, end: g.end, text: orig}}
	case e.final == k:
		return e.finalChunks(k)
	case e.final >= 0 && k > e.final:
		return []chunk{{start: g.start, end: g.end, text: "", rule: e.textRule[k-1]}}
	case g.drop:
		rule := e.textRule[k]
		if rule == "" && k > 0 {
			rule = e.textRule[k-1]
		}
		return []chunk{{start: g.start, end: g.end, text: "", rule: rule}}
	}
	prevNL := k > 0 && endsLine(e.text[k-1])
	if !g.comment {
		return []chunk{e.wsChunk(k, g.start, g.end, e.leadBreaks(g), g.rule, prevNL, g.toComment)}
	}
	var out []chunk
	for i, p := range g.pieces {
		if p.comment {
			if i == 0 && prevNL && e.want[k] != keepLevel {
				out = append(out, chunk{start: p.start, end: p.start, text: e.commentIndent(k), rule: "indentation"})
			}
			out = append(out, chunk{comment: &p, k: k})
			continue
		}
		var breaks int
		rule := ""
		if i == 0 && g.dropLead {
			out = append(out, chunk{start: p.start, end: p.end, rule: "remove_unused_imports"})
			continue
		}
		switch {
		case i == g.lead && (g.nl >= 0 || i < len(g.pieces)-1):
			breaks, rule = e.leadBreaks(g), g.rule
		case i == len(g.pieces)-1:
			breaks, rule = e.tailBreaks(g), g.tailRule
		default:
			breaks = e.cap(lineBreaks(e.src[p.start:p.end]))
		}
		toComment := i < len(g.pieces)-1 || g.toComment
		out = append(out, e.wsChunk(k, p.start, p.end, breaks, rule, prevNL && i == 0, toComment))
	}
	if last := g.pieces[len(g.pieces)-1]; last.comment && g.pieces[g.lead].comment && g.nl > 0 && !e.commentEndsLine(last) {
		// No whitespace to hold the wanted line breaks: they follow the comment.
		out = append(out, chunk{start: g.end, end: g.end, text: strings.Repeat(e.eol, e.cap(g.nl)) + e.indentFor(k, false), rule: g.rule})
	}
	return out
}

// wsChunk renders whitespace [start, end) before unit k (or before a
// comment of its gap when toComment).
func (e *engine) wsChunk(k int, start, end uint32, breaks int, rule string, prevNL, toComment bool) chunk {
	orig := string(e.src[start:end])
	if breaks == 0 {
		text := orig
		if strings.ContainsAny(orig, "\n\r") {
			text = " "
		}
		if prevNL {
			text = e.indentFor(k, toComment)
			if rule == "" {
				rule = "indentation"
			}
			return chunk{start: start, end: end, text: text, rule: rule}
		}
		if !toComment && !e.gaps[k].comment {
			text, rule = e.spaces(k, orig, rule)
		}
		if rule == "" {
			rule = "spacing"
		}
		return chunk{start: start, end: end, text: text, rule: rule}
	}
	ind := e.indentFor(k, toComment)
	if e.want[k] == keepLevel {
		ind = ""
		if i := strings.LastIndexAny(orig, "\n\r"); i >= 0 {
			ind = orig[i+1:]
		} else if !toComment {
			ind = e.lineLead(k)
		}
	}
	text := strings.Repeat(e.eol, breaks) + ind
	if rule == "" || lineBreaks([]byte(orig)) == breaks {
		rule = e.wsRule(orig, text)
	}
	return chunk{start: start, end: end, text: text, rule: rule}
}

// indentFor returns the indentation of a line starting unit k, or a
// comment line before it.
func (e *engine) indentFor(k int, toComment bool) string {
	// After removed statements the line holds the next unit written; a
	// removed keyword (array, list) leaves its line's level to what follows.
	j := k
	for j+1 < len(e.sig) && e.deleted[j] {
		if e.kind(j) == syntax.TSemicolon {
			k = j + 1
		}
		j++
	}
	if toComment {
		if e.want[k] == keepLevel {
			return ""
		}
		return e.commentIndent(k)
	}
	if e.want[k] == keepLevel {
		return e.origIndent(k)
	}
	return e.lineIndent(k)
}

// spaces renders the space of gap k on one line.
func (e *engine) spaces(k int, orig, rule string) (string, string) {
	g := &e.gaps[k]
	want, why := g.sp, rule
	p := k - 1
	for p > 0 && e.deleted[p] {
		p-- // space from the unit written before
	}
	if want == -1 && k > 0 {
		want, why = e.spacing(p, k)
	}
	if why == "" {
		why = "spacing"
	}
	switch want {
	case -1:
		return orig, why
	case atLeastOne:
		if orig != "" && strings.Trim(orig, " \t") == "" {
			return orig, why
		}
		return " ", why
	case 0:
		if orig != "" && !e.spaceOK(p, k) {
			return orig, why
		}
		return "", why
	}
	return strings.Repeat(" ", want), why
}

// wsRule names the setting behind a change of whitespace with line breaks.
func (e *engine) wsRule(orig, text string) string {
	ob, nb := lineBreaks([]byte(orig)), lineBreaks([]byte(text))
	switch {
	case ob != nb:
		return "max_blank_lines" // rules adding breaks name themselves
	case strings.ReplaceAll(strings.ReplaceAll(orig, "\r\n", "\n"), "\r", "\n") != orig &&
		strings.ReplaceAll(strings.ReplaceAll(orig, "\r\n", "\n"), "\r", "\n") == strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n"):
		return "line_ending"
	}
	oi, ni := orig[strings.LastIndexAny(orig, "\n\r")+1:], text[strings.LastIndexAny(text, "\n\r")+1:]
	if oi != ni {
		return "indentation"
	}
	if e.eol != "\n" || strings.Contains(orig, "\r") {
		return "line_ending"
	}
	return "trim_trailing_whitespace"
}

// commentChunk renders comment p before unit k.
func (e *engine) commentChunk(k int, p piece, lineStart bool) chunk {
	orig := string(e.src[p.start:p.end])
	text, rule := orig, ""
	if e.s.HashComment && strings.HasPrefix(text, "#") && !strings.HasPrefix(text, "#[") && (p.start == 0 || e.src[p.start-1] != '/') {
		text, rule = "//"+text[1:], "hash_comment"
	}
	if strings.HasPrefix(text, "/*") && strings.ContainsAny(text, "\n\r") {
		lines := splitLines(text)
		ind := ""
		reindent := lineStart && e.want[k] != keepLevel && starred(lines[1:])
		if reindent {
			ind = e.commentIndent(k)
		}
		for i := 1; i < len(lines); i++ {
			l := lines[i]
			trimmed := strings.TrimLeft(l, " \t")
			if reindent && strings.HasPrefix(trimmed, "*") {
				l = ind + " " + trimmed
			}
			if e.s.TrimTrailingWhitespace {
				l = strings.TrimRight(l, " \t")
			}
			lines[i] = l
		}
		if e.s.TrimTrailingWhitespace {
			lines[0] = strings.TrimRight(lines[0], " \t")
		}
		joined := strings.Join(lines, e.eol)
		if joined != text && rule == "" {
			rule = "indentation"
		}
		text = joined
	} else if e.s.TrimTrailingWhitespace {
		if t := strings.TrimRight(text, " \t"); t != text {
			text = t
			if rule == "" {
				rule = "trim_trailing_whitespace"
			}
		}
	}
	return chunk{start: p.start, end: p.end, text: text, rule: rule}
}

// splitLines splits s at line breaks, dropping them.
func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return strings.Split(s, "\n")
}

// finalChunks renders the gap after the last code unit: its comments, then
// exactly one line break.
func (e *engine) finalChunks(k int) []chunk {
	g := &e.gaps[k]
	eof := len(e.sig) - 1 // what follows is the end of the file
	var out []chunk
	prevNL := k > 0 && endsLine(e.text[k-1])
	last := g.end
	for i, p := range g.pieces {
		if !p.comment {
			if i == 0 && g.dropLead {
				out = append(out, chunk{start: p.start, end: p.end, rule: "remove_unused_imports"})
			}
			continue
		}
		if i > 0 && !g.pieces[i-1].comment && !(i == 1 && g.dropLead) {
			ws := g.pieces[i-1]
			breaks := e.cap(lineBreaks(e.src[ws.start:ws.end]))
			out = append(out, e.wsChunk(eof, ws.start, ws.end, breaks, "", prevNL && i == 1, true))
		}
		out = append(out, chunk{comment: &p, k: eof})
		last = p.end
	}
	start := g.start
	if len(out) > 0 {
		start = last
	}
	text := e.eol
	if prevNL && len(out) == 0 {
		text = ""
	}
	return append(out, chunk{start: start, end: g.end, text: text, rule: "final_newline"})
}

// starred reports whether every non-empty line starts with "*" after its
// indentation, as in a docblock.
func starred(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimLeft(l, " \t")
		if t != "" && !strings.HasPrefix(t, "*") {
			return false
		}
	}
	return true
}

// atLineStart reports whether the output so far ends at the start of a
// line, indentation aside.
func atLineStart(out string) bool {
	t := strings.TrimRight(out, " \t")
	return t == "" || strings.HasSuffix(t, "\n") || strings.HasSuffix(t, "\r")
}
