package stringnormalization

import (
	"strings"
	"unicode"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// stringNormalization reports case conversions applied before trimming or
// cutting (pattern A) and redundant nested case conversions (pattern B).
type stringNormalization struct{}

func (stringNormalization) ID() string { return "StringNormalization" }
func (stringNormalization) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func isLengthFunc(name string) bool {
	switch name {
	case "trim", "ltrim", "rtrim", "substr", "mb_substr":
		return true
	}
	return false
}

func isBasicCaseFunc(name string) bool {
	switch name {
	case "strtolower", "strtoupper", "mb_convert_case", "mb_strtolower", "mb_strtoupper":
		return true
	}
	return false
}

func isCaseFunc(name string) bool {
	switch name {
	case "ucfirst", "lcfirst", "ucwords":
		return true
	}
	return isBasicCaseFunc(name)
}

func (stringNormalization) Check(ctx *analysis.Context, n syntax.Node) {
	outer := n.(*syntax.FuncCall)
	oname := ctx.GlobalFunctionName(outer)
	if !(isLengthFunc(oname) || isCaseFunc(oname)) {
		return
	}
	oargs, ok := astquery.CallArgValues(outer) // D1
	if !ok || len(oargs) == 0 {
		return
	}
	inner, ok := oargs[0].(*syntax.FuncCall) // D2
	if !ok {
		return
	}
	_, written, _ := astquery.CallName(inner)
	iname := ctx.GlobalFunctionName(inner)
	if !isCaseFunc(iname) {
		return
	}
	iargs, ok := astquery.CallArgValues(inner)
	if !ok || len(iargs) == 0 { // E7
		return
	}
	subject := iargs[0]

	if isLengthFunc(oname) { // pattern A
		if !isBasicCaseFunc(iname) || !cutFirstApplies(oname, oargs) { // D3, D4
			return
		}
		cut := oname == "substr" || oname == "mb_substr"
		if cut && iname == "mb_convert_case" && !keepsPositions(ctx, iargs) { // D3a
			return
		}
		if cut && strings.HasPrefix(iname, "mb_") {
			// D3b: a multibyte conversion may change the length and needs
			// whole characters, so the swap is left to the author.
			ctx.Report(outer.Span(), "Cut first, then change the case (check multibyte input by hand).")
			return
		}
		src := ctx.Src
		os, is, ss := outer.Span(), inner.Span(), subject.Span()
		swappedOuter := string(src[os.Start:is.Start]) + ctx.Text(subject) + string(src[is.End:os.End])
		repl := string(src[is.Start:ss.Start]) + swappedOuter + string(src[ss.End:is.End])
		ctx.Report(os, "Cut first, then change the case: '"+repl+"'.", diagnostic.Fix{
			Title: "Swap the calls",
			Edits: func() []diagnostic.TextEdit {
				return []diagnostic.TextEdit{{Span: os, NewText: repl}}
			},
		})
		return
	}

	// pattern B (D5)
	switch {
	case oname == iname: // D5a
		// Repeating a call is only a no-op with the same mode, encoding
		// or delimiters at both levels.
		if len(iargs) != len(oargs) {
			return
		}
		for i := 1; i < len(iargs); i++ {
			if !astquery.EquivalentFoldNames(ctx.File, iargs[i], oargs[i]) {
				return
			}
		}
	case isBasicCaseFunc(oname) && !isBasicCaseFunc(iname) && (iname != "ucwords" || len(iargs) == 1): // D5b
	default:
		return
	}
	is := inner.Span()
	repl := ctx.Text(subject)
	ctx.Report(is, "The inner '"+written+"(...)' call has no effect here.", diagnostic.Fix{
		Title: "Remove the inner call",
		Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: is, NewText: repl}}
		},
	})
}

// keepsPositions reports whether an mb_convert_case call converts every
// character the same way wherever it stands: its mode is one of the
// upper/lower/fold constants. Title case, like ucfirst, depends on what
// precedes a character, and any other expression is unknown.
func keepsPositions(ctx *analysis.Context, iargs []syntax.Expr) bool {
	if len(iargs) < 2 {
		return false
	}
	switch strings.TrimPrefix(ctx.Text(iargs[1]), `\`) {
	case "MB_CASE_UPPER", "MB_CASE_LOWER", "MB_CASE_FOLD",
		"MB_CASE_UPPER_SIMPLE", "MB_CASE_LOWER_SIMPLE", "MB_CASE_FOLD_SIMPLE":
		return true
	}
	return false
}

// cutFirstApplies implements D4 for a length function with a case-converted
// first argument.
func cutFirstApplies(oname string, oargs []syntax.Expr) bool {
	if oname == "substr" || oname == "mb_substr" || len(oargs) == 1 {
		return true
	}
	if len(oargs) != 2 {
		return false
	}
	// Only a literal without interpolation: the characters of "$x" or
	// "{$x}" are unknown.
	lit, ok := oargs[1].(*syntax.Literal)
	if !ok || lit.LitKind != syntax.LitString {
		return false
	}
	raw := strings.TrimLeft(lit.Raw, "bB")
	// single-quoted literals and nowdocs have no escape sequences
	escapes := !strings.HasPrefix(strings.TrimLeft(raw, "< \t"), "'")
	if strings.HasPrefix(raw, "<<<") {
		// heredoc/nowdoc: the body, between the opening and closing lines
		raw = raw[strings.IndexByte(raw, '\n') : strings.LastIndexByte(raw, '\n')+1]
	} else if strings.ContainsAny(raw, "\r\n") {
		return false
	}
	if escapes {
		// double-quoted or heredoc: an octal escape such as "\101" spells a
		// character the letter scan below cannot see
		for i := 0; i+1 < len(raw); i++ {
			if raw[i] == '\\' {
				if raw[i+1] >= '0' && raw[i+1] <= '7' {
					return false
				}
				i++ // `\\` is a backslash, not the start of an escape
			}
		}
	}
	for _, r := range raw {
		if unicode.IsLetter(r) {
			return false
		}
	}
	// a range such as '@..Z' or '!..~' covers letters too
	for i := 1; i+2 < len(raw); i++ {
		if raw[i] == '.' && raw[i+1] == '.' && rangeHasLetter(raw[i-1], raw[i+2]) {
			return false
		}
	}
	return true
}

func rangeHasLetter(lo, hi byte) bool {
	return lo <= 'z' && hi >= 'a' || lo <= 'Z' && hi >= 'A'
}
