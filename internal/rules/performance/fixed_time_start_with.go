package performance

import (
	"strconv"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// fixedTimeStartWith reports `strpos($s, 'lit') === 0` prefix checks and
// suggests strncmp()/strncasecmp() bounded by the needle length.
type fixedTimeStartWith struct{}

func init() { register(fixedTimeStartWith{}) }

func (fixedTimeStartWith) ID() string { return "FixedTimeStartWith" }

func (fixedTimeStartWith) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func (fixedTimeStartWith) Check(ctx *analysis.Context, n syntax.Node) {
	call, name := perfCall(ctx, n, "strpos", "stripos") // D1
	if call == nil {
		return
	}
	b, ok := call.Parent().(*syntax.Binary) // D2
	if !ok || (b.Op.Kind != syntax.TIsIdentical && b.Op.Kind != syntax.TIsNotIdentical) {
		return
	}
	args, ok := perfPlainArgs(call.Args)
	if !ok || len(args) != 2 { // D3
		return
	}
	lit, ok := args[1].(*syntax.Literal) // D4
	if !ok {
		return
	}
	needle, ok := util.StringLiteralValue(lit.Raw)
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
	fn = util.QualifiedBuiltinFor(ctx, fn, call) // F1: keep a global `\`; qualify when a bare name would not reach the builtin
	// F1: byte length of the decoded needle (see spec Divergences).
	repl := fn + "(" + ctx.Text(args[0]) + ", " + ctx.Text(args[1]) + ", " + strconv.Itoa(len(needle)) + ")"
	span := call.Span()
	ctx.Report(span, "Use '"+repl+"' for a length-independent prefix check.", analysis.Fix{
		Title: "Replace with '" + repl + "'",
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: span, NewText: repl}} },
	})
}
