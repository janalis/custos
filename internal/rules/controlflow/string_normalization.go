package controlflow

import (
	"strings"
	"unicode"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// stringNormalization reports case conversions applied before trimming or
// cutting (pattern A) and redundant nested case conversions (pattern B).
type stringNormalization struct{}

func init() { register(stringNormalization{}) }

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
	oargs, ok := util.CallArgValues(outer) // D1
	if !ok || len(oargs) == 0 {
		return
	}
	inner, ok := oargs[0].(*syntax.FuncCall) // D2
	if !ok {
		return
	}
	_, written, _ := util.CallName(inner)
	iname := ctx.GlobalFunctionName(inner)
	if !isCaseFunc(iname) {
		return
	}
	iargs, ok := util.CallArgValues(inner)
	if !ok || len(iargs) == 0 { // E7
		return
	}
	subject := iargs[0]

	if isLengthFunc(oname) { // pattern A
		if !isBasicCaseFunc(iname) || !cutFirstApplies(oname, oargs) { // D3, D4
			return
		}
		src := ctx.Src
		os, is, ss := outer.Span(), inner.Span(), subject.Span()
		swappedOuter := string(src[os.Start:is.Start]) + ctx.Text(subject) + string(src[is.End:os.End])
		repl := string(src[is.Start:ss.Start]) + swappedOuter + string(src[ss.End:is.End])
		ctx.Report(os, "Cut first, then change the case: '"+repl+"'.", analysis.Fix{
			Title: "Swap the calls",
			Edits: func() []analysis.TextEdit {
				return []analysis.TextEdit{{Span: os, NewText: repl}}
			},
		})
		return
	}

	// pattern B (D5)
	switch {
	case oname == iname: // D5a
		if oname == "ucwords" && len(iargs) > 1 && (len(oargs) < 2 || !util.EquivalentFoldNames(ctx.File, iargs[1], oargs[1])) {
			return // the inner call splits words on other delimiters
		}
	case isBasicCaseFunc(oname) && !isBasicCaseFunc(iname) && (iname != "ucwords" || len(iargs) == 1): // D5b
	default:
		return
	}
	is := inner.Span()
	repl := ctx.Text(subject)
	ctx.Report(is, "The inner '"+written+"(...)' call has no effect here.", analysis.Fix{
		Title: "Remove the inner call",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: is, NewText: repl}}
		},
	})
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
	if strings.HasPrefix(raw, "<<<") {
		// heredoc/nowdoc: the body, between the opening and closing lines
		raw = raw[strings.IndexByte(raw, '\n') : strings.LastIndexByte(raw, '\n')+1]
	} else if strings.ContainsAny(raw, "\r\n") {
		return false
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
