package performance

import (
	"strconv"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// subStrShortHandUsage reports substr()/mb_substr() lengths computed as
// `strlen($s) - x` that are a negative constant or unnecessary.
type subStrShortHandUsage struct{}

func init() { register(subStrShortHandUsage{}) }

func (subStrShortHandUsage) ID() string { return "SubStrShortHandUsage" }

func (subStrShortHandUsage) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func (subStrShortHandUsage) Check(ctx *analysis.Context, n syntax.Node) {
	call, name := perfCall(ctx, n, "substr", "mb_substr") // D1
	if call == nil {
		return
	}
	args, ok := perfPlainArgs(call.Args)
	if !ok || len(args) < 3 || len(args) > 4 { // D2
		return
	}
	if len(args) == 4 && name == "substr" { // E1b: substr has no 4th parameter
		return
	}
	subject, start, length := args[0], args[1], args[2]
	sub, ok := length.(*syntax.Binary) // D3
	if !ok || sub.Op.Kind != syntax.TMinus {
		return
	}
	// D4: the length function must count in the same unit as the substr
	// variant (bytes for substr, characters for mb_substr).
	lenName := "strlen"
	if name == "mb_substr" {
		lenName = "mb_strlen"
	}
	strlen, _ := perfCall(ctx, sub.Left, lenName)
	if strlen == nil {
		return
	}
	largs, ok := perfPlainArgs(strlen.Args)
	if !ok || len(largs) != 1 || !util.EquivalentFoldNames(ctx.File, largs[0], subject) {
		return
	}
	span := length.Span()
	if !util.EquivalentFoldNames(ctx.File, sub.Right, start) { // D5
		s, ok1 := decimalInt(ctx, start)
		r, ok2 := decimalInt(ctx, sub.Right)
		if !ok1 || !ok2 || s < 0 { // D6 (negative start: spec Divergences)
			return
		}
		if d := s - r; d < 0 {
			if d < -2 {
				// For a string shorter than R the original length is
				// negative and still keeps characters; `d` keeps none.
				// Some length L in ((R+start)/2, R) exists once d <= -3.
				return
			}
			repl := strconv.FormatInt(d, 10)
			ctx.Report(span, "Pass '"+repl+"' as the length instead.", replaceFix(span, repl))
			return
		}
	}
	// Drop (F2/F3).
	first, last := args[0].Span(), args[len(args)-1].Span()
	repl := ctx.Text(subject) + ", " + ctx.Text(start)
	if len(args) == 4 {
		repl += ", null, " + ctx.Text(args[3])
	}
	list := syntax.Span{Start: first.Start, End: last.End}
	ctx.Report(span, "The length '"+ctx.Text(length)+"' is unnecessary; remove it.", analysis.Fix{
		Title: "Remove the length",
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: list, NewText: repl}} },
	})
}

// decimalInt parses an integer literal, optionally negated, whose source text
// is a plain 32-bit decimal number (`- 2`, `0x2`, `1_0`, `1.0` fail).
func decimalInt(ctx *analysis.Context, e syntax.Expr) (int64, bool) {
	lit := e
	if u, ok := e.(*syntax.Unary); ok && u.Op.Kind == syntax.TMinus {
		lit = u.Expr
	}
	if l, ok := lit.(*syntax.Literal); !ok || l.LitKind != syntax.LitInt {
		return 0, false
	}
	v, err := strconv.ParseInt(ctx.Text(e), 10, 32)
	return v, err == nil
}
