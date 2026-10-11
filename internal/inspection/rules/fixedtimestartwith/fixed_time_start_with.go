package fixedtimestartwith

import (
	"strconv"
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// fixedTimeStartWith reports `strpos($s, 'lit') === 0` prefix checks and
// suggests strncmp()/strncasecmp() bounded by the needle length.
type fixedTimeStartWith struct{}

func (fixedTimeStartWith) ID() string               { return "FixedTimeStartWith" }
func (fixedTimeStartWith) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (fixedTimeStartWith) Check(ctx *analysis.Context, n syntax.Node) {
	call, name := semanticquery.GlobalCall(ctx, n, "strpos", "stripos") // D1
	if call == nil {
		return
	}
	b, ok := call.Parent().(*syntax.Binary) // D2
	if !ok || (b.Op.Kind != syntax.TIsIdentical && b.Op.Kind != syntax.TIsNotIdentical) {
		return
	}
	args, ok := astquery.ArgValues(call.Args)
	if !ok || len(args) != 2 { // D3
		return
	}
	lit, ok := args[1].(*syntax.Literal) // D4
	if !ok {
		return
	}
	needle, ok := astquery.StringLiteralValue(lowerHexEscapes(lit.Raw))
	if !ok || needle == "" { // D4a: strpos() with '' was false before PHP 8
		return
	}
	other := b.Left
	if other == syntax.Expr(call) {
		other = b.Right
	}
	if ctx.Text(other) != "0" { // D5
		return
	}
	fn := "strncmp"
	if name == "stripos" {
		fn = "strncasecmp"
	}
	fn = semanticquery.QualifiedBuiltinFor(ctx, fn, call) // F1: keep a global `\`; qualify when a bare name would not reach the builtin
	// F1: byte length of the decoded needle (see spec Divergences).
	repl := fn + "(" + ctx.Text(args[0]) + ", " + ctx.Text(args[1]) + ", " + strconv.Itoa(len(needle)) + ")"
	span := call.Span()
	ctx.Report(span, "Use '"+repl+"' for a length-independent prefix check.", diagnostic.Fix{
		Title: "Replace with '" + repl + "'",
		Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: span, NewText: repl}} },
	})
}

// lowerHexEscapes rewrites the escape introducer `\X` as `\x` in a quoted
// literal's source: PHP decodes both alike in double quotes, and neither in
// single quotes, so the decoded length is unchanged while the decoder only
// has to know the lowercase form. Backslashes are consumed in pairs, so the
// `X` of `\\X41` is left alone.
func lowerHexEscapes(raw string) string {
	if !strings.Contains(raw, `\X`) {
		return raw
	}
	var b strings.Builder
	b.Grow(len(raw))
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		if c == '\\' && i+1 < len(raw) {
			b.WriteByte(c)
			i++
			if c = raw[i]; c == 'X' {
				c = 'x'
			}
		}
		b.WriteByte(c)
	}
	return b.String()
}
