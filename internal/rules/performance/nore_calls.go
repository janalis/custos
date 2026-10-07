package performance

import (
	"regexp"
	"strconv"
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/meta"
	"custos/internal/syntax"
)

// NotOptimalRegularExpressions: call shape checks and plain string API
// replacements (D20–D23).

var (
	noreTextForm  = regexp.MustCompile(`^(\^)?([A-Za-z0-9_-]+|\\[.*+?])(\$)?$`)
	noreTrimStart = regexp.MustCompile(`^\^(\\s|[^.])[+*]$`)
	noreTrimEnd   = regexp.MustCompile(`^(\\s|[^.])[+*]\$$`)
	noreTrimBoth  = regexp.MustCompile(`^\^(\\s|[^.])[+*]\|(\\s|[^.])[+*]\$$`)
	noreRegexSyn  = regexp.MustCompile(`[^\\][\^$.*+?\\\[\](){}!|-]|\\[dDhHsSvVwWRb]`)
	noreUnescape  = regexp.MustCompile(`\\([.+*?-])`)
)

var noreCaseConversions = []string{"strtolower", "strtoupper", "mb_strtolower", "mb_strtoupper"}

func (c *noreCase) checkCall() {
	ctx, call := c.ctx, c.call
	args, ok := perfPlainArgs(call.Args)
	if !ok {
		args = nil
		for _, a := range perfArgs(call.Args) {
			args = append(args, a.Value)
		}
	}
	argc := len(call.Args.Args)

	if c.fn == "preg_quote" && argc == 1 { // D20
		ctx.ReportSeverity(call.Span(), meta.SeverityWarning, "Pass the delimiter to preg_quote() so it is escaped as well.")
	}
	if c.fn == "preg_match_all" && argc == 2 && util.IsLogicalOperand(call) { // D21
		ctx.ReportSeverity(call.Name.Span(), meta.SeverityInfo, "Use preg_match() when only testing for a match.")
	}
	if argc >= 2 && len(args) == argc && c.body != "" { // D22
		c.plainAPI(args, argc)
	}
	if c.fn == "preg_match" && argc == 2 && len(args) == 2 { // D23
		if inner, _ := perfCall(ctx, args[1], noreCaseConversions...); inner != nil {
			msg := "Drop the case conversion and add the /i flag instead."
			if strings.IndexByte(c.mods, 'i') >= 0 {
				msg = "The case conversion is redundant: the pattern is already case-insensitive."
			}
			ctx.ReportSeverity(inner.Span(), meta.SeverityWarning, msg)
		}
	}
}

func noreUnescapeText(s string) string { return noreUnescape.ReplaceAllString(s, "$1") }

// plainAPI implements D22.
func (c *noreCase) plainAPI(args []syntax.Expr, argc int) {
	ctx, call := c.ctx, c.call
	norm := noreNormalize(c.body)
	ci := strings.IndexByte(c.mods, 'i') >= 0
	// bi spells an inserted builtin so a namespaced or imported function of
	// that name does not capture it.
	bi := func(name string) string { return util.QualifiedBuiltin(ctx, name, call.Span().Start) }
	x1 := ctx.Text(args[1])
	x2 := ""
	if argc > 2 {
		x2 = ctx.Text(args[2])
	}

	if m := noreTextForm.FindStringSubmatch(norm); m != nil {
		start, end := m[1] != "", m[3] != ""
		t := noreUnescapeText(m[2])
		site, inverted, siteOK := noreSite(call)
		anchored := strings.IndexByte(c.mods, 'A') >= 0 // `A` anchors at the start
		// Without `D`, `$` also matches before a final newline, so `^T$`
		// accepts "T\n" and is not the same as an identity comparison.
		strictEnd := strings.IndexByte(c.mods, 'D') >= 0
		perLine := strings.IndexByte(c.mods, 'm') >= 0 && (start || end) // ^/$ match at line breaks
		find := bi("strpos")
		if ci {
			find = bi("stripos")
		}
		repl := ""
		switch {
		case anchored || perLine:
		case c.fn == "preg_match" && !siteOK:
		case c.fn == "preg_match" && argc == 2 && start && end && !ci && strictEnd: // D22a
			op := "==="
			if inverted {
				op = "!=="
			}
			repl = `"` + t + `" ` + op + " " + x1
		case c.fn == "preg_match" && argc == 2 && start && !end: // D22b
			op := "==="
			if inverted {
				op = "!=="
			}
			repl = "0 " + op + " " + find + "(" + x1 + `, "` + t + `")`
		case c.fn == "preg_match" && argc == 2 && !start && !end: // D22c
			op := "!=="
			if inverted {
				op = "==="
			}
			repl = "false " + op + " " + find + "(" + x1 + `, "` + t + `")`
		case c.fn == "preg_replace" && argc == 3 && !start && !end: // D22d
			g := bi("str_replace")
			if ci {
				g = bi("str_ireplace")
			}
			repl = g + `("` + t + `", ` + x1 + ", " + x2 + ")"
		}
		if repl != "" {
			if c.fn != "preg_match" {
				site = call // D22d replaces the call by a string function
			} else if needsParensAsComparison(site) {
				noreReplaceText(ctx, site.Span(), repl, "("+repl+")")
				return
			}
			noreReplace(ctx, site.Span(), repl)
			return
		}
	}

	if c.fn == "preg_replace" && argc == 3 { // D22e
		if lit, ok := args[1].(*syntax.Literal); ok && lit.LitKind == syntax.LitString && len(ctx.Text(lit)) == 2 {
			ch, ok := "", false
			if m := noreTrimStart.FindStringSubmatch(norm); m != nil {
				ch, ok = m[1], true
			} else if m := noreTrimEnd.FindStringSubmatch(norm); m != nil {
				ch, ok = m[1], true
			} else if m := noreTrimBoth.FindStringSubmatch(norm); m != nil && m[1] == m[2] {
				ch, ok = m[1], true
			}
			if ok {
				if !noreTrimModsOK(c.mods, ch) || noreMetaChars(ch) {
					return
				}
				if strings.HasSuffix(c.body, "$") && ch != `\s` && strings.IndexByte(c.mods, 'D') < 0 {
					// `c+$` without `D` stops before a final newline, which
					// rtrim/trim would not; `\s+$` consumes it, so it is fine.
					return
				}
				h := "trim"
				switch {
				case !strings.HasPrefix(c.body, "^"):
					h = "rtrim"
				case !strings.HasSuffix(c.body, "$"):
					h = "ltrim"
				}
				h = bi(h)
				// PCRE's \s is space, \t, \n, \v, \f and \r; trim()'s default
				// list differs (\0 instead of \f), so it is spelled out.
				repl := h + "(" + x2 + `, " \t\n\r\v\f")`
				if ch != `\s` {
					repl = h + "(" + x2 + ", '" + noreUnescapeText(ch) + "')"
				}
				noreReplace(ctx, call.Span(), repl)
				return
			}
		}
	}

	if c.fn == "preg_split" && (argc == 2 || argc == 3) && c.mods == "" { // D22f
		if u, ok := noreExplodeDelimiter(norm); ok && (argc == 2 || norePositiveInt(args[2])) {
			repl := bi("explode") + `("` + strings.ReplaceAll(u, `"`, `\"`) + `", ` + x1
			if argc == 3 {
				repl += ", " + x2
			}
			noreReplace(ctx, call.Span(), repl+")")
		}
	}
}

// noreTrimModsOK reports whether the modifiers keep a trim pattern on
// character ch equivalent to trim(): only D and S are neutral; i is fine
// unless ch has a case variant (`/^a+/i` also strips A); m, u, U (lazy `+`
// strips one character), x (whitespace ignored) and the rest are not.
func noreTrimModsOK(mods, ch string) bool {
	for _, m := range mods {
		switch {
		case m == 'D' || m == 'S':
		case m == 'i' && strings.ToLower(ch) == strings.ToUpper(ch):
		default:
			return false
		}
	}
	return true
}

// noreMetaChars reports whether a one-character trim/split body is a regex
// metacharacter (or a quote/backslash that cannot be written as is in the
// replacement), so it is not the literal character.
func noreMetaChars(ch string) bool {
	return len(ch) == 1 && strings.Contains(`\^$.[]|()?*+{'"`, ch)
}

// noreExplodeDelimiter returns the literal string a D22f body splits on:
// a single literal character, `[c]` with c literal, or a body without regex
// syntax (escapes `\.` `\+` `\*` `\?` `\-` resolved). Bodies starting with
// a metacharacter, or keeping another escape (`\/`, `\n`), are refused.
func noreExplodeDelimiter(norm string) (string, bool) {
	if len(norm) == 3 && norm[0] == '[' && norm[2] == ']' && !strings.Contains(`\^]`, norm[1:2]) {
		return norm[1:2], true // a class of one character, `.` included
	}
	if len(norm) == 1 {
		return norm, !noreMetaChars(norm) || norm == `"` || norm == "'"
	}
	if noreRegexSyn.MatchString(norm) || strings.ContainsAny(norm[:1], `^$.*+?[](){}|`) {
		return "", false
	}
	u := noreUnescapeText(norm)
	return u, !strings.Contains(u, `\`)
}

// norePositiveInt reports whether e is a positive integer literal (a limit
// that preg_split() and explode() treat alike).
func norePositiveInt(e syntax.Expr) bool {
	lit, ok := e.(*syntax.Literal)
	if !ok || lit.LitKind != syntax.LitInt {
		return false
	}
	n, err := strconv.ParseInt(strings.ReplaceAll(lit.Raw, "_", ""), 0, 64)
	return err == nil && n > 0
}

func noreReplace(ctx *analysis.Context, span syntax.Span, repl string) {
	noreReplaceText(ctx, span, repl, repl)
}

// noreReplaceText reports repl and fixes with text (repl, parenthesised
// where the context binds tighter than the replacement).
func noreReplaceText(ctx *analysis.Context, span syntax.Span, repl, text string) {
	ctx.ReportSeverity(span, meta.SeverityWarning, "Replace with '"+repl+"'.", analysis.Fix{
		Title: "Replace with '" + repl + "'",
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: span, NewText: text}} },
	})
}

// noreSite implements the D22 context and inversion of a preg_match()
// call whose 0/1 result is replaced by a boolean test (custos diverges from
// upstream, see spec): ok is false when no replacement keeps the meaning.
//   - logical operand: the call, or its direct `!` (inverted);
//   - direct operand of a comparison with a number literal: the comparison,
//     inverted when it holds for 0 and not for 1; skipped when it holds for
//     both or neither (`== 2`, `< 5`);
//   - any other operator operand (arithmetic, `@`, casts, a comparison
//     with something else or through parentheses): skipped;
//   - a value position (assignment, return, argument…): the call.
func noreSite(call *syntax.FuncCall) (site syntax.Node, inverted, ok bool) {
	if util.IsLogicalOperand(call) {
		if u, isNot := call.Parent().(*syntax.Unary); isNot && u.Op.Kind == syntax.TExclaim {
			return u, true, true
		}
		return call, false, true
	}
	parent, _ := util.ParentSkipParens(call)
	switch parent.(type) {
	case *syntax.Binary:
	case *syntax.Unary, *syntax.Instanceof:
		return nil, false, false
	default:
		return call, false, true
	}
	b, direct := call.Parent().(*syntax.Binary)
	if !direct {
		return nil, false, false // `(preg_match(…)) === 0`
	}
	op, other := b.Op.Kind, b.Left
	if other == syntax.Expr(call) {
		other = b.Right
	} else { // number on the left: `1 > C` reads `C < 1`
		switch op {
		case syntax.TLess:
			op = syntax.TGreater
		case syntax.TGreater:
			op = syntax.TLess
		case syntax.TIsSmallerOrEqual:
			op = syntax.TIsGreaterOrEqual
		case syntax.TIsGreaterOrEqual:
			op = syntax.TIsSmallerOrEqual
		}
	}
	holds0, ok0 := noreCompare(op, 0, other)
	holds1, ok1 := noreCompare(op, 1, other)
	if !ok0 || !ok1 || holds0 == holds1 {
		return nil, false, false
	}
	return b, holds0, true
}

// noreCompare evaluates `v op n` for a number literal n (optionally negated).
func noreCompare(op syntax.TokenKind, v int64, n syntax.Expr) (holds, ok bool) {
	neg := false
	if u, isNeg := n.(*syntax.Unary); isNeg && u.Op.Kind == syntax.TMinus {
		neg, n = true, u.Expr
	}
	lit, isLit := n.(*syntax.Literal)
	if !isLit {
		return false, false
	}
	raw := strings.ReplaceAll(lit.Raw, "_", "")
	var f float64
	isInt := lit.LitKind == syntax.LitInt
	switch {
	case isInt:
		i, err := strconv.ParseInt(raw, 0, 64)
		if err != nil {
			return false, false
		}
		f = float64(i)
	case lit.LitKind == syntax.LitFloat:
		x, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return false, false
		}
		f = x
	default:
		return false, false
	}
	if neg {
		f = -f
	}
	x := float64(v)
	switch op {
	case syntax.TIsEqual:
		return x == f, true
	case syntax.TIsNotEqual:
		return x != f, true
	case syntax.TIsIdentical: // an int is never identical to a float
		return isInt && x == f, true
	case syntax.TIsNotIdentical:
		return !isInt || x != f, true
	case syntax.TLess:
		return x < f, true
	case syntax.TGreater:
		return x > f, true
	case syntax.TIsSmallerOrEqual:
		return x <= f, true
	case syntax.TIsGreaterOrEqual:
		return x >= f, true
	}
	return false, false
}
