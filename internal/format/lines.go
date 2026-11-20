package format

import (
	"slices"

	"github.com/janalis/custos/internal/php/syntax"
)

// unitAt returns the unit starting at off, or -1.
func (e *engine) unitAt(off uint32) int {
	i := e.at(off)
	if i < 0 {
		return -1
	}
	return e.unitOf[i]
}

// unitAfter returns the first unit starting at or after off, or -1.
func (e *engine) unitAfter(off uint32) int {
	for i := e.index(off); i < len(e.toks); i++ {
		if u := e.unitOf[i]; u >= 0 {
			return u
		}
	}
	return -1
}

// unitBefore returns the last unit ending at or before off, or -1.
func (e *engine) unitBefore(off uint32) int {
	for i := e.index(off) - 1; i >= 0; i-- {
		if u := e.unitOf[i]; u >= 0 {
			return u
		}
	}
	return -1
}

// lines decides the line structure: braces, blank lines and statements.
func (e *engine) lines() {
	s := e.s
	for k := 1; k < len(e.sig); k++ {
		e.statementPerLine(k)
	}
	for k := range e.sig {
		if e.kind(k) == syntax.TLBrace {
			e.braceLines(k)
		}
	}
	for _, st := range e.f.Stmts {
		e.walkLines(st)
	}
	e.topLevel(e.f.Stmts)
	if s.ElsePosition != "keep" || s.CatchPosition != "keep" {
		e.continuations()
	}
}

// statementPerLine starts a statement on its own line.
func (e *engine) statementPerLine(k int) {
	if !e.s.StatementPerLine || e.gaps[k].comment || e.gaps[k].frozen {
		return
	}
	p := e.kind(k - 1)
	if !e.start[e.sig[k]] || e.leadBreaks(&e.gaps[k]) > 0 {
		return
	}
	if p != syntax.TSemicolon && p != syntax.TRBrace {
		return
	}
	switch e.kind(k) {
	case syntax.TElse, syntax.TElseif, syntax.TCatch, syntax.TFinally, syntax.TWhile:
		if p == syntax.TRBrace {
			return // continues the statement the brace closes
		}
	}
	if e.role[e.sig[k-1]] == roleForSemicolon || e.inList(k) || e.inOneLineClosure(k) {
		return
	}
	e.setNL(k, 1, "statement_per_line")
}

// inList reports whether unit k sits directly in a list bracket.
func (e *engine) inList(k int) bool {
	depth := 0
	for j := k - 1; j >= 0; j-- {
		switch kd := e.kind(j); {
		case closer(kd):
			depth++
		case opener(kd):
			if depth == 0 {
				b := e.br[e.sig[j]]
				return b.list() || b == brControl || b == brGroup || b == brHooks
			}
			depth--
		}
	}
	return false
}

// braceStyle returns the setting placing the opening brace of kind b.
func (e *engine) braceStyle(b bracket) (string, string) {
	switch b {
	case brClass:
		return e.s.ClassBrace, "class_brace"
	case brFunction:
		return e.s.FunctionBrace, "function_brace"
	case brClosure:
		return e.s.ClosureBrace, "closure_brace"
	case brAnonClass:
		return e.s.AnonymousClassBrace, "anonymous_class_brace"
	case brControlBlock, brSwitch, brNamespace:
		return e.s.ControlBrace, "control_brace"
	case brMatch:
		if e.s.ControlBrace == "keep" {
			return "keep", ""
		}
		return "same_line", "control_brace"
	}
	return "keep", ""
}

// braceLines places the opening brace k and the lines of its body.
func (e *engine) braceLines(k int) {
	b := e.br[e.sig[k]]
	c := e.match[k]
	empty := c == k+1 && !e.gaps[c].comment
	style, rule := e.braceStyle(b)
	if b == brControlBlock && !e.controlBody(k) {
		style = "keep"
	}
	if b == brFunction && e.multilineSignature(k) && style != "keep" {
		style = "same_line"
	}
	if empty && e.s.EmptyBody == "inline" && (b == brClass || b == brFunction || b == brClosure || b == brAnonClass) {
		e.setNL(k, 0, "empty_body")
		e.setSpace(c, 0, "empty_body")
		return
	}
	switch {
	case style == "same_line" && !e.gaps[k].comment:
		e.setNL(k, 0, rule)
	case style == "next_line" && !e.gaps[k].comment && !(empty && (b == brClosure || b == brAnonClass)):
		e.setNL(k, 1, rule)
	}
	if empty && style == "next_line" && (b == brClass || b == brFunction) && !e.gaps[c].comment {
		e.setNL(c, 1, rule) // {} after a declaration: the closing brace gets its own line
	}
	if empty || !b.block() {
		return
	}
	// The body starts and ends on lines of its own.
	oneLine := e.leadBreaks(&e.gaps[k+1]) == 0 && !e.gaps[k+1].comment
	if oneLine && (b == brClosure || b == brAnonClass) && e.s.SingleLineClosures && e.sameLine(k, c) {
		return
	}
	if style != "keep" {
		if oneLine {
			e.setNL(k+1, 1, "brace_body")
		}
		if e.leadBreaks(&e.gaps[c]) == 0 && !e.gaps[c].comment {
			e.setNL(c, 1, "brace_body")
		}
	}
	if b == brClass || b == brAnonClass {
		e.blankAfterOpen(k+1, e.s.ClassOpenBlankLine, "class_open_blank_line")
		e.blankBeforeClose(c, e.s.ClassCloseBlankLine, "class_close_blank_line")
		return
	}
	if e.s.TrimBlockBlankLines {
		e.blankAfterOpen(k+1, "none", "trim_block_blank_lines")
		e.blankBeforeClose(c, "none", "trim_block_blank_lines")
	}
}

// controlBody reports whether the block opened at k is the body of a
// control structure (not a statement block of its own).
func (e *engine) controlBody(k int) bool {
	p := k - 1
	if p < 0 {
		return false
	}
	switch e.kind(p) {
	case syntax.TRParen, syntax.TElse, syntax.TTry, syntax.TDo, syntax.TFinally:
		return true
	}
	return false
}

// sameLine reports whether units a and b are on one line in the source.
func (e *engine) sameLine(a, b int) bool {
	for k := a + 1; k <= b; k++ {
		if lineBreaks(e.src[e.gaps[k].start:e.gaps[k].end]) > 0 {
			return false
		}
	}
	return true
}

// multilineSignature reports whether the parameter list before the
// function body at k ends with its closing parenthesis on a line of its own.
func (e *engine) multilineSignature(k int) bool {
	for j := k - 1; j >= 0; j-- {
		if e.kind(j) == syntax.TLParen && e.br[e.sig[j]] == brParams {
			return e.startsLine(e.match[j])
		}
		if e.kind(j) == syntax.TLBrace || e.kind(j) == syntax.TSemicolon {
			return false
		}
	}
	return false
}

func blankCount(v string) int {
	if v == "one" {
		return 2
	}
	return 1
}

// blankAfterOpen sets the blank line after an opening brace (gap k).
func (e *engine) blankAfterOpen(k int, v, rule string) {
	if v == "keep" || e.leadBreaks(&e.gaps[k]) == 0 {
		return
	}
	e.setNL(k, blankCount(v), rule)
}

// blankBeforeClose sets the blank line before a closing brace (gap k).
func (e *engine) blankBeforeClose(k int, v, rule string) {
	g := &e.gaps[k]
	if v == "keep" {
		return
	}
	if g.comment {
		if g.tailNL > 0 {
			e.setTail(k, blankCount(v), rule)
		}
		return
	}
	if e.leadBreaks(g) > 0 {
		e.setNL(k, blankCount(v), rule)
	}
}

// continuations places else, elseif, catch, finally and the while of a
// do-while loop relative to the closing brace before them.
func (e *engine) continuations() {
	for k := 1; k < len(e.sig); k++ {
		if e.kind(k-1) != syntax.TRBrace || e.gaps[k].comment {
			continue
		}
		var v, rule string
		switch e.kind(k) {
		case syntax.TElse, syntax.TElseif:
			v, rule = e.s.ElsePosition, "else_position"
		case syntax.TCatch, syntax.TFinally:
			v, rule = e.s.CatchPosition, "catch_position"
		case syntax.TWhile:
			if e.br[e.sig[e.match[k-1]]] == brControlBlock && e.kind(e.match[k-1]-1) == syntax.TDo {
				v, rule = "same_line", "control_brace"
			}
		}
		switch v {
		case "same_line":
			e.setNL(k, 0, rule)
		case "next_line":
			e.setNL(k, 1, rule)
		}
	}
}

// walkLines applies the blank-line rules of statement lists and classes.
func (e *engine) walkLines(n syntax.Node) {
	if isNil(n) || e.inString(n.Span().Start) {
		return
	}
	switch v := n.(type) {
	case *syntax.ClassLike:
		e.members(v.Members)
	case *syntax.Block:
		e.blankBefore(v.Stmts)
	case *syntax.Case:
		e.blankBefore(v.Stmts)
	case *syntax.Switch:
		if e.s.CaseSeparation {
			for i := 1; i < len(v.Cases); i++ {
				if len(v.Cases[i-1].Stmts) > 0 {
					e.atLeastBlank(e.unitAt(v.Cases[i].Span().Start), "case_separation")
				}
			}
		}
	case *syntax.Namespace:
		e.topLevel(v.Stmts)
	case *syntax.ExprStmt:
		if _, ok := v.Expr.(*syntax.Throw); ok && e.s.SingleLineThrow {
			e.singleLine(v)
		}
	}
	syntax.Children(n, e.walkLines)
}

func memberKind(n syntax.Stmt) string {
	switch n.(type) {
	case *syntax.Method:
		return "method"
	case *syntax.Property:
		return "property"
	case *syntax.ClassConst:
		return "constant"
	case *syntax.TraitUse:
		return "trait"
	case *syntax.EnumCase:
		return "case"
	}
	return ""
}

// members separates class members with blank lines.
func (e *engine) members(list []syntax.Stmt) {
	for i := 1; i < len(list); i++ {
		k := e.unitAt(list[i].Span().Start)
		prev, cur := memberKind(list[i-1]), memberKind(list[i])
		var v, rule string
		switch {
		case prev == "trait" && cur != "trait" && cur != "" && e.s.BlankLineAfterTraitUse:
			v, rule = "one", "blank_line_after_trait_use"
		case cur == "method" || (prev == "method" && cur != ""):
			v, rule = e.s.MethodSeparation, "method_separation"
		case cur == "property" && prev == "property":
			v, rule = e.s.PropertySeparation, "property_separation"
		case cur == "constant" && prev == "constant":
			v, rule = e.s.ConstantSeparation, "constant_separation"
		case cur == "property" || cur == "constant":
			if s := e.s.PropertySeparation; cur == "property" && s == "one" {
				v, rule = s, "property_separation"
			} else if s := e.s.ConstantSeparation; cur == "constant" && s == "one" {
				v, rule = s, "constant_separation"
			}
		}
		if v == "" || v == "keep" || e.leadBreaks(&e.gaps[k]) == 0 && !e.gaps[k].comment {
			continue
		}
		e.setNL(k, blankCount(v), rule)
	}
}

func stmtKeyword(n syntax.Stmt) string {
	switch v := n.(type) {
	case *syntax.Return:
		return "return"
	case *syntax.Break:
		return "break"
	case *syntax.Continue:
		return "continue"
	case *syntax.If:
		return "if"
	case *syntax.For:
		return "for"
	case *syntax.Foreach:
		return "foreach"
	case *syntax.While:
		return "while"
	case *syntax.DoWhile:
		return "do"
	case *syntax.Switch:
		return "switch"
	case *syntax.Try:
		return "try"
	case *syntax.Declare:
		return "declare"
	case *syntax.ExprStmt:
		switch x := v.Expr.(type) {
		case *syntax.Throw:
			return "throw"
		case *syntax.Yield, *syntax.YieldFrom:
			return "yield"
		case *syntax.Assign:
			if _, ok := x.Value.(*syntax.Yield); ok {
				return "yield"
			}
		}
	}
	return ""
}

// blankBefore puts a blank line before the listed statements.
func (e *engine) blankBefore(list []syntax.Stmt) {
	if len(e.s.BlankLineBefore) == 0 {
		return
	}
	for i := 1; i < len(list); i++ {
		kw := stmtKeyword(list[i])
		if kw == "" || !slices.Contains(e.s.BlankLineBefore, kw) {
			continue
		}
		k := e.unitAt(list[i].Span().Start)
		if k < 1 || e.leadBreaks(&e.gaps[k]) == 0 && !e.gaps[k].comment {
			continue
		}
		if kw == "yield" && stmtKeyword(list[i-1]) == "yield" {
			continue
		}
		if e.leadBreaks(&e.gaps[k]) < 2 {
			e.setNL(k, 2, "blank_line_before")
		}
	}
}

// topLevel applies the header rules to a list of file or namespace
// statements: blank lines after namespace and imports, import groups.
func (e *engine) topLevel(all []syntax.Stmt) {
	s := e.s
	list := make([]syntax.Stmt, 0, len(all))
	for _, st := range all {
		if k := e.unitAt(st.Span().Start); k < 0 || !e.deleted[k] {
			list = append(list, st) // removed imports are gone
		}
	}
	for i, st := range list {
		if v, ok := st.(*syntax.Namespace); ok {
			if !v.Braced && s.BlankLineAfterNamespace {
				next := e.nextUnit(v)
				if len(v.Stmts) > 0 {
					next = e.unitAt(v.Stmts[0].Span().Start)
				}
				if next > 0 && e.kind(next) != syntax.TEOF {
					e.atLeastBlank(next, "blank_line_after_namespace")
				}
			}
			continue
		}
		next := e.nextUnit(st)
		if next < 0 || e.kind(next) == syntax.TEOF || e.kind(next) == syntax.TRBrace {
			continue
		}
		v, ok := st.(*syntax.Use)
		if !ok {
			continue
		}
		var nextUse *syntax.Use
		if i+1 < len(list) {
			nextUse, _ = list[i+1].(*syntax.Use)
		}
		switch {
		case nextUse == nil && s.BlankLineAfterImports:
			// Right after the import, before any comment that follows it.
			e.atLeastBlank(e.unitAfter(v.Span().End), "blank_line_after_imports")
		case nextUse != nil && s.BlankLineBetweenImportGroups && e.leadBreaks(&e.gaps[e.written(next)]) > 0:
			if useKind(v) != useKind(nextUse) {
				e.setNL(next, 2, "blank_line_between_import_groups")
			} else if !e.gaps[e.written(next)].comment {
				e.setNL(next, 1, "blank_line_between_import_groups")
			}
		}
	}
}

func useKind(u *syntax.Use) syntax.UseKind {
	if u.Type == syntax.UseNormal && len(u.Items) > 0 && u.Prefix == nil {
		return u.Items[0].Type
	}
	return u.Type
}

// nextUnit returns the first unit after statement st.
func (e *engine) nextUnit(st syntax.Node) int {
	k := e.unitAfter(st.Span().End)
	for k > 0 && k < len(e.sig)-1 && e.deleted[k] {
		k++ // past removed statements
	}
	return k
}

// atLeastBlank puts at least one blank line before unit k.
func (e *engine) atLeastBlank(k int, rule string) {
	if k < 1 {
		return
	}
	k = e.written(k)
	if e.leadBreaks(&e.gaps[k]) < 2 {
		e.setNL(k, 2, rule)
	}
}

// inOneLineClosure reports whether unit k is in the body of a closure kept
// on one line.
func (e *engine) inOneLineClosure(k int) bool {
	if !e.s.SingleLineClosures {
		return false
	}
	depth := 0
	for j := k - 1; j >= 0; j-- {
		switch kd := e.kind(j); {
		case closer(kd):
			depth++
		case opener(kd) && depth > 0:
			depth--
		case opener(kd):
			b := e.br[e.sig[j]]
			return (b == brClosure || b == brAnonClass) && e.sameLine(j, e.match[j])
		}
	}
	return false
}

// singleLine writes statement st on one line when nothing in it needs a line
// of its own: no comment, brace or heredoc.
func (e *engine) singleLine(st syntax.Stmt) {
	sp := st.Span()
	a, b := e.unitAt(sp.Start), e.unitBefore(sp.End)
	for k := a; k <= b; k++ {
		switch e.kind(k) {
		case syntax.TLBrace, syntax.TStartHeredoc:
			return
		}
		if k > a && (e.gaps[k].comment || e.gaps[k].frozen) {
			return
		}
	}
	for k := a + 1; k <= b; k++ {
		if e.leadBreaks(&e.gaps[k]) > 0 {
			e.setNL(k, 0, "single_line_throw")
			e.gaps[k].sp = -1
		}
	}
	e.noWrap = append(e.noWrap, region{sp.Start, sp.End})
}
