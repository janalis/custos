// Package format lays out PHP source code: indentation, spaces, line
// breaks, braces, blank lines and the token spellings a style guide fixes.
//
// One pass over the token stream decides the whitespace between every two
// significant tokens, using the syntax tree for context. The pass yields
// the formatted output and the list of changes that produce it, so a check
// reports exactly what a fix writes.
package format

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
)

// Change is one edit of the source: [Start, End) becomes New.
type Change struct {
	Rule  string // the setting behind the edit
	Start uint32
	End   uint32
	New   string
	// Report marks a finding formatting cannot fix: it edits nothing.
	Report bool
	// Group, when not 0, links edits that only work together (the two
	// brackets of an array, a rebuilt import block): apply all or none.
	Group int
}

// Message returns the description of the change.
func (c Change) Message() string {
	if m, ok := messages[c.Rule]; ok {
		return m
	}
	return "Fix the layout"
}

// Result is the outcome of formatting one file.
type Result struct {
	Output  []byte
	Changes []Change
	// Skipped explains why the file was left as written.
	Skipped string
}

// Changed reports whether formatting changes the file or reports
// something on it.
func (r Result) Changed() bool { return len(r.Changes) > 0 }

// Options configure Source.
type Options struct {
	Settings *Settings
	PHP      phpversion.Version
}

// Source formats src, the content of the PHP file path.
func Source(path string, src []byte, opt Options) Result {
	res, again := sourceOnce(path, src, opt)
	for pass := 0; again && pass < 2 && res.Skipped == ""; pass++ {
		// A rewrite joined tokens into one unit (a swapped comparison, a
		// rebuilt import block): lay the result out once more.
		next, more := sourceOnce(path, res.Output, opt)
		if next.Skipped != "" {
			break
		}
		res = Result{Output: next.Output, Changes: compose(res.Changes, next.Changes, next.Output)}
		again = more
	}
	return res
}

// sourceOnce formats src in one pass; again reports that a rewrite merged
// units, so another pass may lay out what they hide.
func sourceOnce(path string, src []byte, opt Options) (Result, bool) {
	s := opt.Settings
	if s == nil {
		r, _ := Resolve(DefaultPreset, false)
		s = &r.Settings
	}
	php := opt.PHP
	if php == 0 {
		php = phpversion.Default
	}
	if ignoreFile.Match(src) {
		return Result{Output: src, Skipped: "ignored by a format-ignore comment"}, false
	}
	lex := php
	f := syntax.Parse(path, src, syntax.Options{Version: php})
	if len(f.Errors) > 0 {
		// Code newer than the configured version: parse it, and lex the
		// output, as the newest version reads it.
		if g := syntax.Parse(path, src, syntax.Options{Version: phpversion.Max, Permissive: true}); len(g.Errors) < len(f.Errors) {
			f, lex = g, phpversion.Max
		}
	}
	if n := len(f.Errors); n > 0 {
		return Result{Output: src, Skipped: plural(n, "syntax error")}, false
	}
	toks := f.Tokens
	if n := len(toks); n == 0 || toks[n-1].Kind != syntax.TEOF {
		end := uint32(len(src))
		toks = append(toks[:n:n], syntax.Token{Kind: syntax.TEOF, Start: end, End: end})
	}
	e := &engine{src: src, f: f, s: s, php: php, lex: lex, toks: toks, final: -1}
	e.annotate()
	e.units()
	e.pair()
	e.prepare()
	e.tokens()
	for _, st := range f.Stmts {
		e.syntaxRewrites(st)
	}
	e.risky()
	e.imports()
	e.fileEnds()
	if !e.template {
		e.lines()
		e.operatorLinebreaks()
		e.wrapPartial()
		e.signatureBraces()
	}
	e.indentation()
	if !e.template {
		e.layout()
	}
	e.trailingCommas()
	out, changes := e.emit()
	changes = append(changes, e.longLineReports(out)...)
	if string(out) == string(src) {
		return Result{Output: src, Changes: changes}, false
	}
	if why := checkOutput(e, out); why != "" {
		return Result{Output: src, Skipped: "formatting would change the code (" + why + "); please report it"}, false
	}
	return Result{Output: out, Changes: changes}, e.merged
}

// checkOutput is the safety check of every pass; tests replace it.
var checkOutput = (*engine).verify

func plural(n int, what string) string {
	if n == 1 {
		return "1 " + what
	}
	return fmt.Sprintf("%d %ss", n, what)
}

// prepare builds the per-unit state.
func (e *engine) prepare() {
	n := len(e.sig)
	e.unitOf = make([]int, len(e.toks))
	for i := range e.unitOf {
		e.unitOf[i] = -1
	}
	for k, t := range e.sig {
		e.unitOf[t] = k
	}
	e.text = make([]string, n)
	e.textRule = make([]string, n)
	e.deleted = make([]bool, n)
	e.group = make([]int, n)
	for k := range e.sig {
		e.text[k] = e.orig(k)
	}
	e.buildGaps()
	e.template = e.isTemplate()
	e.eol = e.lineEnding()
	e.ignoreRegions()
	for k := 1; k < len(e.gaps); k++ {
		if e.kind(k-1) == syntax.THaltCompilerData {
			e.gaps[k].frozen = true // nothing follows the data but the end of file
		}
	}
	if e.template {
		for k := range e.gaps {
			// Markup keeps its surroundings.
			if k > 0 && isMarkup(e.kind(k-1)) || isMarkup(e.kind(k)) {
				e.gaps[k].frozen = true
			}
		}
	}
}

func isMarkup(k syntax.TokenKind) bool {
	switch k {
	case syntax.TInlineHTML, syntax.TOpenTag, syntax.TOpenTagWithEcho, syntax.TCloseTag, syntax.THaltCompilerData:
		return true
	}
	return false
}

// isTemplate reports whether the file mixes PHP with markup: anything
// besides one leading `<?php`, a byte order mark, a leading shebang line and
// whitespace after a final `?>`.
func (e *engine) isTemplate() bool {
	for k := range e.sig {
		// A second opening tag follows a closing one, which ends the check.
		switch e.kind(k) {
		case syntax.TOpenTagWithEcho:
			return true
		case syntax.TInlineHTML:
			txt := e.orig(k)
			switch {
			case k == 0 && (txt == "\xEF\xBB\xBF" || strings.HasPrefix(txt, "#!") && strings.Count(txt, "\n") <= 1):
			case k == len(e.sig)-2 && strings.TrimSpace(txt) == "":
			default:
				return true
			}
		case syntax.TCloseTag:
			if k < len(e.sig)-3 || (k == len(e.sig)-3 && e.kind(k+1) != syntax.TInlineHTML) {
				return true
			}
		}
	}
	return false
}

// lineEnding returns the line break to write.
func (e *engine) lineEnding() string {
	switch e.s.LineEnding {
	case "crlf":
		return "\r\n"
	case "cr":
		return "\r"
	case "lf":
		return "\n"
	}
	for i, c := range e.src {
		switch c {
		case '\n':
			return "\n"
		case '\r':
			if i+1 < len(e.src) && e.src[i+1] == '\n' {
				return "\r\n"
			}
			return "\r"
		}
	}
	return "\n"
}

var (
	ignoreFile  = regexp.MustCompile(`@(?:custos-format-ignore|mago-format(?:ter)?-ignore)(?:[^-\w]|$)`)
	ignoreStart = regexp.MustCompile(`@(?:custos-format-ignore-start|mago-format(?:ter)?-ignore-start|formatter:off)\b`)
	ignoreEnd   = regexp.MustCompile(`@(?:custos-format-ignore-end|mago-format(?:ter)?-ignore-end|formatter:on)\b`)
	ignoreNext  = regexp.MustCompile(`@(?:custos-format-ignore-next|mago-format(?:ter)?-ignore-next)\b`)
)

// ignoreRegions records the code the ignore comments protect.
func (e *engine) ignoreRegions() {
	var open uint32
	opened := false
	for _, t := range e.toks {
		if t.Kind != syntax.TComment && t.Kind != syntax.TDocComment {
			continue
		}
		c := e.src[t.Start:t.End]
		switch {
		case ignoreStart.Match(c) && !opened:
			open, opened = t.Start, true
		case ignoreEnd.Match(c) && opened:
			e.ignored = append(e.ignored, region{open, t.End})
			opened = false
		case ignoreNext.Match(c):
			if k := e.unitAfter(t.End); k >= 0 {
				if n := e.statementAt(e.toks[e.sig[k]].Start); n != nil {
					e.ignored = append(e.ignored, region{t.Start, n.Span().End})
				}
			}
		}
	}
	if opened {
		e.ignored = append(e.ignored, region{open, uint32(len(e.src))})
	}
	for k := range e.gaps {
		g := &e.gaps[k]
		if e.inRegions(e.ignored, g.start) && g.end > g.start && e.inRegions(e.ignored, g.end-1) {
			g.frozen = true
		}
	}
}

// statementAt returns the statement or class member starting at off.
func (e *engine) statementAt(off uint32) syntax.Node {
	var found syntax.Node
	var visit func(n syntax.Node)
	visit = func(n syntax.Node) {
		if found != nil || isNil(n) {
			return
		}
		sp := n.Span()
		if sp.End <= off || sp.Start > off {
			return
		}
		if _, ok := n.(syntax.Stmt); ok && sp.Start == off {
			found = n
			return
		}
		syntax.Children(n, visit)
	}
	for _, st := range e.f.Stmts {
		visit(st)
	}
	return found
}

// verify checks that out holds the same code as the source: the same
// significant tokens apart from the rewrites, and no syntax error.
func (e *engine) verify(out []byte) string {
	got := significant(out, e.lex)
	var want []syntax.Token
	var wantText []string
	var batch strings.Builder
	flush := func() {
		if batch.Len() == 0 {
			return
		}
		src := "<?php\n" + batch.String()
		for _, t := range significant([]byte(src), e.lex) {
			if t.Kind == syntax.TOpenTag || t.Kind == syntax.TEOF || t.Kind == syntax.THaltCompilerData {
				continue // the data after __halt_compiler is a unit of its own
			}
			want = append(want, t)
			wantText = append(wantText, src[t.Start:t.End])
		}
		batch.Reset()
	}
	for k := range e.sig {
		txt := e.text[k]
		if txt == "" {
			continue
		}
		if e.kind(k) == syntax.TOpenTag && strings.TrimSpace(txt) != strings.TrimSpace(e.orig(k)) {
			flush()
			for _, t := range significant([]byte(txt), e.lex) {
				want = append(want, t)
				wantText = append(wantText, txt[t.Start:t.End])
			}
			continue
		}
		if isMarkup(e.kind(k)) {
			flush()
			want = append(want, syntax.Token{Kind: e.kind(k)})
			wantText = append(wantText, txt)
			continue
		}
		batch.WriteString(txt)
		batch.WriteByte('\n')
	}
	flush()
	if len(got) != len(want) {
		return fmt.Sprintf("%d tokens instead of %d", len(got), len(want))
	}
	for i, t := range got {
		if t.Kind != want[i].Kind {
			return fmt.Sprintf("token %d is %v instead of %v", i, t.Kind, want[i].Kind)
		}
		switch t.Kind {
		case syntax.TOpenTag, syntax.TCloseTag, syntax.TEOF:
			continue
		}
		if string(out[t.Start:t.End]) != wantText[i] {
			return fmt.Sprintf("token %d reads %q instead of %q", i, out[t.Start:t.End], wantText[i])
		}
	}
	if f := syntax.Parse("", out, syntax.Options{Version: e.lex, Permissive: true}); len(f.Errors) > 0 {
		if g := syntax.Parse("", e.src, syntax.Options{Version: e.lex, Permissive: true}); len(g.Errors) < len(f.Errors) {
			return "syntax error"
		}
	}
	return ""
}

// significant lexes src and returns its non-trivia tokens.
func significant(src []byte, php phpversion.Version) []syntax.Token {
	toks, _ := syntax.Lex(src, syntax.LexOptions{Version: php})
	out := toks[:0]
	for _, t := range toks {
		if !t.Kind.IsTrivia() {
			out = append(out, t)
		}
	}
	return out
}
