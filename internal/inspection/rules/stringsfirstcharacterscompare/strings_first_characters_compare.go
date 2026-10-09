package stringsfirstcharacterscompare

import (
	"math"
	"strconv"
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// stringsFirstCharactersCompare reports strncmp()/strncasecmp() calls whose
// literal length argument does not match the length of the literal prefix.
type stringsFirstCharactersCompare struct{}

func (stringsFirstCharactersCompare) ID() string { return "StringsFirstCharactersCompare" }
func (stringsFirstCharactersCompare) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func (stringsFirstCharactersCompare) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if name := ctx.GlobalFunctionName(call); name != "strncmp" && name != "strncasecmp" { // D1
		return
	}
	if call.Args == nil || len(call.Args.Args) != 3 { // D2
		return
	}
	var vals [3]syntax.Expr
	for i, a := range call.Args.Args {
		arg, ok := a.(*syntax.Arg)
		if !ok || arg.Unpack || arg.Value == nil {
			return
		}
		vals[i] = arg.Value
	}
	if !astquery.IsNumberLiteral(vals[2]) { // D3
		return
	}
	lit, ok := astquery.QuotedStringValue(vals[1]) // D4
	if !ok {
		if lit, ok = astquery.QuotedStringValue(vals[0]); !ok {
			return
		}
	}
	l := len(lit) // D5: bytes of the decoded value (see spec divergences)
	lenText := ctx.Text(vals[2])
	num, ok := parsePHPInt32(lenText) // D6
	if !ok || l == 0 || l == num {    // D7
		return
	}
	span := vals[2].Span()
	want := strconv.Itoa(l)
	ctx.Report(span, "Length "+lenText+" does not match the "+want+"-character literal.", diagnostic.Fix{
		Title: "Use the literal's length",
		Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: span, NewText: want}}
		},
	})
}

// parsePHPInt32 parses an integer literal as PHP reads it (decimal, octal,
// hex, binary, `_` separators), optionally preceded directly by `-`, within
// the signed 32-bit range.
func parsePHPInt32(s string) (int, bool) {
	neg := strings.HasPrefix(s, "-")
	v, ok := astquery.ParseIntLiteral(strings.TrimPrefix(s, "-"))
	if !ok {
		return 0, false
	}
	if neg {
		v = -v
	}
	if v < math.MinInt32 || v > math.MaxInt32 {
		return 0, false
	}
	return int(v), true
}
