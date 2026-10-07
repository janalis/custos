package performance

import (
	"regexp"
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/meta"
	"custos/internal/syntax"
)

// NotOptimalRegularExpressions: call shape checks and plain string API
// replacements (D20–D23).

var (
	noreTextForm   = regexp.MustCompile(`^(\^)?([A-Za-z0-9_-]+|\\[.*+?])(\$)?$`)
	noreTrimStart  = regexp.MustCompile(`^\^(\\s|[^.])[+*]$`)
	noreTrimEnd    = regexp.MustCompile(`^(\\s|[^.])[+*]\$$`)
	noreTrimBoth   = regexp.MustCompile(`^\^(\\s|[^.])[+*]\|(\\s|[^.])[+*]\$$`)
	noreSplitClass = regexp.MustCompile(`^(?:[^.]|\[[^.]\])$`)
	noreRegexSyn   = regexp.MustCompile(`[^\\][\^$.*+?\\\[\](){}!|-]|\\[dDhHsSvVwWRb]`)
	noreUnescape   = regexp.MustCompile(`\\([.+*?-])`)
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
		inverted := noreInverted(call)
		find := bi("strpos")
		if ci {
			find = bi("stripos")
		}
		repl := ""
		switch {
		case c.fn == "preg_match" && argc == 2 && start && end && !ci: // D22a
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
			noreReplace(ctx, noreContext(call).Span(), repl)
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
				if strings.ContainsAny(c.mods, "mu") {
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
				repl := h + "(" + x2 + ")"
				if ch != `\s` {
					repl = h + "(" + x2 + ", '" + noreUnescapeText(ch) + "')"
				}
				noreReplace(ctx, call.Span(), repl)
				return
			}
		}
	}

	if c.fn == "preg_split" && (argc == 2 || argc == 3) && c.mods == "" { // D22f
		if noreSplitClass.MatchString(norm) || !noreRegexSyn.MatchString(norm) {
			repl := bi("explode") + `("` + noreUnescapeText(norm) + `", ` + x1
			if argc == 3 {
				repl += ", " + x2
			}
			noreReplace(ctx, call.Span(), repl+")")
		}
	}
}

func noreReplace(ctx *analysis.Context, span syntax.Span, repl string) {
	ctx.ReportSeverity(span, meta.SeverityWarning, "Replace with '"+repl+"'.", analysis.Fix{
		Title: "Replace with '" + repl + "'",
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: span, NewText: repl}} },
	})
}

// noreNumberOperand returns the text of a number literal, optionally
// negated, or ok=false.
func noreNumberOperand(e syntax.Expr) (string, bool) {
	if u, ok := e.(*syntax.Unary); ok && u.Op.Kind == syntax.TMinus {
		if lit, ok := u.Expr.(*syntax.Literal); ok && (lit.LitKind == syntax.LitInt || lit.LitKind == syntax.LitFloat) {
			return "-" + lit.Raw, true
		}
		return "", false
	}
	if lit, ok := e.(*syntax.Literal); ok && (lit.LitKind == syntax.LitInt || lit.LitKind == syntax.LitFloat) {
		return lit.Raw, true
	}
	return "", false
}

// noreCompared returns the binary comparison directly containing call with a
// number on the other side, its operator kind and the number text.
func noreCompared(call *syntax.FuncCall) (*syntax.Binary, syntax.TokenKind, string) {
	b, ok := call.Parent().(*syntax.Binary)
	if !ok {
		return nil, 0, ""
	}
	other := b.Left
	if other == syntax.Expr(call) {
		other = b.Right
	}
	num, ok := noreNumberOperand(other)
	if !ok {
		return nil, 0, ""
	}
	return b, b.Op.Kind, num
}

// noreInverted implements the D22 "inverted" definition.
func noreInverted(call *syntax.FuncCall) bool {
	if util.IsLogicalOperand(call) {
		u, ok := call.Parent().(*syntax.Unary)
		return ok && u.Op.Kind == syntax.TExclaim
	}
	b, op, num := noreCompared(call)
	if b == nil {
		return false
	}
	switch op {
	case syntax.TLess:
		return num == "1"
	case syntax.TIsEqual, syntax.TIsIdentical:
		return num == "0"
	case syntax.TIsNotEqual, syntax.TIsNotIdentical:
		return num == "1"
	}
	return false
}

// noreContext implements the D22 "context" definition.
func noreContext(call *syntax.FuncCall) syntax.Node {
	if util.IsLogicalOperand(call) {
		if u, ok := call.Parent().(*syntax.Unary); ok && u.Op.Kind == syntax.TExclaim {
			return u
		}
	}
	if b, op, _ := noreCompared(call); b != nil {
		switch op {
		case syntax.TIsEqual, syntax.TIsNotEqual, syntax.TIsIdentical, syntax.TIsNotIdentical,
			syntax.TLess, syntax.TGreater, syntax.TIsSmallerOrEqual, syntax.TIsGreaterOrEqual:
			return b
		}
	}
	return call
}
