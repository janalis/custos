package format

import (
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/janalis/custos/internal/php/syntax"
)

// operatorLinebreaks moves operators that end a line to the start of the
// next one.
func (e *engine) operatorLinebreaks() {
	mode := e.s.OperatorLinebreak
	if mode == "keep" {
		return
	}
	for k := 1; k+1 < len(e.sig); k++ {
		r := e.role[e.sig[k]]
		switch {
		case r == roleBoolean:
		case mode == "start" && (r == roleBinary || r == roleConcat || r == roleTernaryQ || r == roleTernaryColon || r == roleAssign):
			if e.kind(k) == syntax.TInstanceof {
				continue
			}
		default:
			continue
		}
		if r == roleTernaryQ && e.role[e.sig[k+1]] == roleTernaryColon {
			continue // ?: moves with its colon
		}
		op := k // where the operator starts
		if r == roleTernaryColon && e.role[e.sig[k-1]] == roleTernaryQ {
			op = k - 1
			if e.leadBreaks(&e.gaps[k]) > 0 || e.gaps[k].comment {
				continue
			}
		}
		before, after := &e.gaps[op], &e.gaps[k+1]
		if before.comment || after.comment || before.frozen || after.frozen {
			continue
		}
		n := e.leadBreaks(after)
		if n == 0 || e.leadBreaks(before) > 0 {
			continue
		}
		e.setNL(op, n, "operator_linebreak")
		e.setNL(k+1, 0, "operator_linebreak")
		after.sp = -1
	}
}

// items returns the units starting the items of the list opened at o.
func (e *engine) items(o int) []int {
	c := e.match[o]
	var out []int
	next := o + 1
	for k := o + 1; k < c; k++ {
		if k == next {
			if e.kind(k) == syntax.TComma {
				return nil // empty slot
			}
			out = append(out, k)
		}
		switch kd := e.kind(k); {
		case opener(kd) && e.match[k] > k:
			k = e.match[k]
		case kd == syntax.TComma:
			next = k + 1
		}
	}
	return out
}

// splittable reports whether the list opened at o may be laid out one
// item per line: a list bracket holding items, with no comment between them.
func (e *engine) splittable(o int) bool {
	b := e.br[e.sig[o]]
	if !b.list() || b == brGroupUse || e.match[o] <= o+1 || e.frozenUnit(o) || e.inRegions(e.noWrap, e.toks[e.sig[o]].Start) {
		return false
	}
	for k := o + 1; k <= e.match[o]; k++ {
		if e.gaps[k].comment || e.gaps[k].frozen {
			return false
		}
	}
	return len(e.items(o)) > 0
}

// split puts every item of the list opened at o, and its closing bracket,
// on a line of its own.
func (e *engine) split(o int, rule string) bool {
	changed := false
	for _, it := range append(e.items(o), e.match[o]) {
		if e.leadBreaks(&e.gaps[it]) == 0 {
			e.setNL(it, 1, rule)
			changed = true
		}
	}
	return changed
}

// partlySplit reports whether some item of the list at o starts a line
// while the list is not one item per line.
func (e *engine) partlySplit(o int) bool {
	its := e.items(o)
	broken, all := false, true
	for _, it := range append(its, e.match[o]) {
		if e.leadBreaks(&e.gaps[it]) > 0 {
			broken = true
		} else {
			all = false
		}
	}
	return broken && !all
}

// wrapPartial normalizes lists split over several lines to one item per line.
func (e *engine) wrapPartial() {
	if e.s.Wrap != "long" {
		return
	}
	for o := range e.sig {
		switch {
		case opener(e.kind(o)) && e.splittable(o) && e.partlySplit(o):
			e.split(o, "wrap")
		case e.kind(o) == syntax.TLParen && e.br[e.sig[o]] == brControl:
			e.splitControl(o)
		}
	}
}

// splitControl gives a condition split over several lines lines of its
// own: the first after the opening parenthesis, the closing one alone.
func (e *engine) splitControl(o int) {
	c := e.match[o]
	broken := false
	for k := o + 1; k <= c; k++ {
		if e.gaps[k].comment || e.gaps[k].frozen {
			return
		}
		if opener(e.kind(k)) && e.match[k] > k {
			k = e.match[k] // inner brackets may span lines
			continue
		}
		broken = broken || e.leadBreaks(&e.gaps[k]) > 0
	}
	if !broken || c <= o+1 {
		return
	}
	e.setNL(o+1, 1, "wrap")
	e.setNL(c, 1, "wrap")
}

// width returns the display width of s.
func textWidth(s string, tab int) int {
	w := 0
	for _, r := range s {
		if r == '\t' {
			w += tab - w%max(tab, 1)
			continue
		}
		w++
	}
	if w == 0 {
		return utf8.RuneCountInString(s)
	}
	return w
}

// longLines returns, for each output line wider than the line length, the
// first unit on it.
func (e *engine) longLines(out []byte) []int {
	limit := e.s.LineLength
	var firsts []int
	k := 0
	start := 0
	for i := 0; i <= len(out); i++ {
		if i < len(out) && out[i] != '\n' {
			continue
		}
		line := string(out[start:i])
		line = strings.TrimRight(line, "\r")
		for k < len(e.sig) && e.outPos[k] < start {
			k++
		}
		if textWidth(line, e.s.TabWidth) > limit && k < len(e.sig) && e.outPos[k] < i && !e.inHeredocLine(k, start) {
			firsts = append(firsts, k)
		}
		start = i + 1
	}
	return firsts
}

// inHeredocLine reports whether the line starting at output offset start
// begins inside a string unit.
func (e *engine) inHeredocLine(k, start int) bool {
	return k > 0 && e.outPos[k-1] < start && e.outPos[k-1]+len(e.text[k-1]) > start
}

// lineUnits returns the units on the output line of unit first.
func (e *engine) lineUnits(first int) []int {
	out := []int{first}
	for k := first + 1; k < len(e.sig) && !e.startsLine(k); k++ {
		out = append(out, k)
	}
	return out
}

// wrapLong splits lists on lines exceeding the line length, outermost
// first, and reports whether anything changed.
func (e *engine) wrapLong(out []byte) bool {
	changed := false
	for _, first := range e.longLines(out) {
		units := e.lineUnits(first)
		// The widest list on the line shortens it most; a short one, such
		// as the single argument of isset(), hardly helps.
		best, width := -1, 0
		for _, k := range units {
			if e.kind(k) == syntax.TLParen && e.br[e.sig[k]] == brControl && e.splitCondition(k) {
				best = -2
				break
			}
			if opener(e.kind(k)) && e.splittable(k) && !e.alreadySplit(k) {
				if w := e.outPos[e.match[k]] - e.outPos[k]; w > width {
					best, width = k, w
				}
			}
		}
		switch {
		case best == -2:
			changed = true
		case best >= 0 && e.split(best, "wrap"):
			changed = true
		}
	}
	return changed
}

// alreadySplit reports whether every item of the list at o starts a line.
func (e *engine) alreadySplit(o int) bool {
	for _, it := range append(e.items(o), e.match[o]) {
		if e.leadBreaks(&e.gaps[it]) == 0 {
			return false
		}
	}
	return true
}

// splitCondition puts a condition of boolean operands on lines of its own,
// one operand per line, operators first.
func (e *engine) splitCondition(o int) bool {
	c := e.match[o]
	var ops []int
	for k := o + 1; k < c; k++ {
		if e.gaps[k].comment || e.gaps[k].frozen {
			return false
		}
		switch kd := e.kind(k); {
		case opener(kd) && e.match[k] > k:
			k = e.match[k]
		case e.role[e.sig[k]] == roleBoolean:
			ops = append(ops, k)
		}
	}
	if len(ops) == 0 || e.leadBreaks(&e.gaps[o+1]) > 0 {
		return false
	}
	e.setNL(o+1, 1, "wrap")
	for _, op := range ops {
		e.setNL(op, 1, "wrap")
		e.setNL(op+1, 0, "wrap")
		e.gaps[op+1].sp = -1
	}
	e.setNL(c, 1, "wrap")
	return true
}

// joinFitting writes a list split one item per line on one line when it
// fits within the line length.
func (e *engine) joinFitting(out []byte) bool {
	changed := false
	for o := range e.sig {
		if !opener(e.kind(o)) || !e.splittable(o) || !e.alreadySplit(o) {
			continue
		}
		c := e.match[o]
		nested := false
		its := e.items(o)
		for k := o + 1; k < c; k++ {
			if e.startsLine(k) && !slices.Contains(its, k) {
				nested = true
				break
			}
		}
		if nested {
			continue
		}
		width := e.joinedWidth(o, c, out)
		if width > e.s.LineLength {
			continue
		}
		for _, it := range append(e.items(o), c) {
			e.setNL(it, 0, "join_fitting_lists")
			e.gaps[it].sp = -1
		}
		if e.kind(c-1) == syntax.TComma && !e.deleted[c-1] {
			e.remove(c-1, "join_fitting_lists")
		}
		changed = true
	}
	return changed
}

// joinedWidth estimates the width of the line holding o once the list up
// to c is joined.
func (e *engine) joinedWidth(o, c int, out []byte) int {
	lineStart := e.outPos[o]
	for lineStart > 0 && out[lineStart-1] != '\n' {
		lineStart--
	}
	w := textWidth(string(out[lineStart:e.outPos[o]]), e.s.TabWidth) + len(e.text[o])
	for k := o + 1; k <= c; k++ {
		if e.kind(k) == syntax.TComma && k == c-1 {
			continue
		}
		sp, _ := e.spacing(k-1, k)
		w += max(sp, 0) + utf8.RuneCountInString(e.text[k])
	}
	for k := c + 1; k < len(e.sig) && !e.startsLine(k); k++ {
		sp, _ := e.spacing(k-1, k)
		w += max(sp, 0) + utf8.RuneCountInString(e.text[k])
	}
	return w
}

// layout wraps long lines and joins fitting lists, rendering the file to
// measure lines after each round.
func (e *engine) layout() {
	if e.s.Wrap == "long" && e.s.LineLength > 0 {
		for range 8 {
			e.trailingCommas()
			out, _ := e.emit()
			if !e.wrapLong(out) {
				break
			}
			e.signatureBraces()
			e.indentation()
		}
	}
	if e.s.ParameterAttributeLine && e.parameterAttributes() {
		e.indentation()
	}
	if e.s.JoinFittingLists && e.s.LineLength > 0 {
		for range 8 { // joining an inner list can make the outer one fit
			out, _ := e.emit()
			if !e.joinFitting(out) {
				break
			}
			e.indentation()
		}
	}
}

// longLineReports reports the lines still wider than the line length.
func (e *engine) longLineReports(out []byte) []Change {
	if !e.s.ReportLongLines || e.template || e.s.LineLength <= 0 {
		return nil
	}
	var reports []Change
	for _, k := range e.longLines(out) {
		off := e.toks[e.sig[k]].Start
		reports = append(reports, Change{Rule: "line_length", Start: off, End: off, Report: true})
	}
	return reports
}

// signatureBraces places the opening brace of functions again once their
// signature may have been split.
func (e *engine) signatureBraces() {
	if e.s.FunctionBrace == "keep" {
		return
	}
	for k := range e.sig {
		if e.kind(k) != syntax.TLBrace || e.br[e.sig[k]] != brFunction || e.gaps[k].comment || e.match[k] == k+1 && !e.gaps[k+1].comment && e.s.EmptyBody == "inline" {
			continue
		}
		switch {
		case e.multilineSignature(k):
			e.setNL(k, 0, "function_brace")
		case e.s.FunctionBrace == "next_line":
			e.setNL(k, 1, "function_brace")
		}
	}
}

// parameterAttributes puts each attribute of a parameter on a line of its
// own in parameter lists written one parameter per line.
func (e *engine) parameterAttributes() bool {
	changed := false
	for o := range e.sig {
		if e.kind(o) != syntax.TLParen || e.br[e.sig[o]] != brParams || e.match[o] <= o+1 || !e.alreadySplit(o) {
			continue
		}
		for k := o + 1; k < e.match[o]; k++ {
			if e.role[e.sig[k]] == roleAttrClose && !e.gaps[k+1].comment && e.leadBreaks(&e.gaps[k+1]) == 0 {
				e.setNL(k+1, 1, "parameter_attribute_line")
				changed = true
			}
		}
	}
	return changed
}
