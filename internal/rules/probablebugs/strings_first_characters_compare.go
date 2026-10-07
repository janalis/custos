package probablebugs

import (
	"math"
	"strconv"
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// stringsFirstCharactersCompare reports strncmp()/strncasecmp() calls whose
// literal length argument does not match the length of the literal prefix.
type stringsFirstCharactersCompare struct{}

func init() { register(stringsFirstCharactersCompare{}) }

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
	if !isIntLiteralArg(vals[2]) { // D3
		return
	}
	lit, ok := quotedStringValue(vals[1]) // D4
	if !ok {
		if lit, ok = quotedStringValue(vals[0]); !ok {
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
	ctx.Report(span, "Length "+lenText+" does not match the "+want+"-character literal.", analysis.Fix{
		Title: "Use the literal's length",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: span, NewText: want}}
		},
	})
}

// isIntLiteralArg reports whether e is a number literal or `-` applied
// directly to one.
func isIntLiteralArg(e syntax.Expr) bool {
	if u, ok := e.(*syntax.Unary); ok && u.Op.Kind == syntax.TMinus {
		e = u.Expr
	}
	l, ok := e.(*syntax.Literal)
	return ok && (l.LitKind == syntax.LitInt || l.LitKind == syntax.LitFloat)
}

// parsePHPInt32 parses an integer literal as PHP reads it (decimal, octal,
// hex, binary, `_` separators), optionally preceded directly by `-`, within
// the signed 32-bit range.
func parsePHPInt32(s string) (int, bool) {
	neg := strings.HasPrefix(s, "-")
	v, ok := util.ParseIntLiteral(strings.TrimPrefix(s, "-"))
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

// quotedStringValue decodes a single- or double-quoted, non-interpolated
// string literal.
func quotedStringValue(e syntax.Expr) (string, bool) {
	l, ok := e.(*syntax.Literal)
	if !ok || l.LitKind != syntax.LitString {
		return "", false
	}
	return util.StringLiteralValue(l.Raw)
}
