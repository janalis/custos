package redundantelseclause

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// redundantElseClause reports the `else`/`elseif` following a braced `if`
// body that always leaves the current flow.
type redundantElseClause struct{}

const (
	redundantElseMsg   = "Drop the 'else' and move its code after the 'if'."
	redundantElseIfMsg = "Turn this 'elseif' into a separate 'if'."
)

func (redundantElseClause) ID() string               { return "RedundantElseClause" }
func (redundantElseClause) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KIf} }
func (redundantElseClause) Check(ctx *analysis.Context, n syntax.Node) {
	ifs := n.(*syntax.If)
	if ifs.Span().Len() == 0 || ifs.Alt || ifs.Cond == nil {
		return
	}
	if _, ok := ifs.Parent().(*syntax.Else); ok { // D1
		return
	}
	body, ok := astquery.BracedBlock(ctx.Src, ifs.Body) // D2
	if !ok {
		return
	}
	// D3/D4
	var alt syntax.Node
	var altBody syntax.Stmt
	isElseIf := false
	switch {
	case len(ifs.ElseIfs) > 0:
		alt, altBody, isElseIf = ifs.ElseIfs[0], ifs.ElseIfs[0].Body, true
		if _, ok := astquery.BracedBlock(ctx.Src, altBody); !ok {
			return
		}
	case ifs.Else != nil:
		alt, altBody = ifs.Else, ifs.Else.Body
		if _, isIf := altBody.(*syntax.If); !isIf {
			if _, ok := astquery.BracedBlock(ctx.Src, altBody); !ok {
				return
			}
		}
	default:
		return
	}
	if !leavesFlow(ctx, body) { // D5
		return
	}
	msg := redundantElseMsg
	if isElseIf {
		msg = redundantElseIfMsg
	}
	kw := astquery.KeywordSpan(ctx.File, alt) // `else` / `elseif`
	src := ctx.Src
	text := func(s syntax.Span) string { return string(src[s.Start:s.End]) }
	bodyEnd := body.Span().End
	// Comments between `}` and the keyword are kept before the moved code.
	// Moved code starts on a new line at the `if` statement's indentation.
	nl := "\n" + astquery.LineIndent(src, ifs.Span().Start)
	gap := strings.TrimSpace(text(syntax.Span{Start: bodyEnd, End: kw.Start}))
	if gap != "" {
		gap = nl + gap
	}
	if !redundantElseMovable(ctx, ifs, altBody) {
		ctx.Report(kw, msg)
		return
	}
	ctx.Report(kw, msg, diagnostic.Fix{
		Title: "Remove the redundant clause",
		Edits: func() []diagnostic.TextEdit {
			if isElseIf {
				ei := ifs.ElseIfs[0]
				is := ifs.Span()
				var b strings.Builder
				b.WriteString("if (")
				b.WriteString(ctx.Text(ifs.Cond))
				b.WriteString(") ")
				b.WriteString(ctx.Text(body))
				b.WriteString(gap)
				b.WriteString(nl)
				b.WriteString(text(syntax.Span{Start: is.Start, End: ifs.Cond.Span().Start}))
				b.WriteString(ctx.Text(ei.Cond))
				b.WriteString(text(syntax.Span{Start: ifs.Cond.Span().End, End: body.Span().Start}))
				b.WriteString(ctx.Text(ei.Body))
				b.WriteString(text(syntax.Span{Start: ei.Span().End, End: is.End}))
				return []diagnostic.TextEdit{{Span: is, NewText: b.String()}}
			}
			removed := syntax.Span{Start: bodyEnd, End: alt.Span().End}
			var moved string
			if nested, isIf := altBody.(*syntax.If); isIf { // F2
				moved = nl + ctx.Text(nested)
			} else { // F1
				bs := altBody.Span()
				inner := strings.TrimSpace(text(syntax.Span{Start: bs.Start + 1, End: bs.End - 1}))
				switch {
				case inner == "":
				case strings.HasPrefix(inner, ";") && gap == "":
					moved = inner
				default:
					moved = nl + inner
				}
			}
			return []diagnostic.TextEdit{{Span: removed, NewText: gap + moved}}
		},
	})
}

// redundantElseMovable reports whether the clause's code can move after the
// if (custos): the if must sit in a statement list — as the unbraced body of
// another if or a loop the moved code would run unconditionally — and the
// clause must declare no function or class, which PHP would then hoist
// (a conditional declaration becoming unconditional).
func redundantElseMovable(ctx *analysis.Context, ifs *syntax.If, altBody syntax.Stmt) bool {
	if _, _, ok := astquery.StmtList(ctx.File, ifs); !ok {
		return false
	}
	blk, ok := altBody.(*syntax.Block)
	if !ok {
		return true // an `else if` moves as one statement
	}
	for _, st := range blk.Stmts {
		switch st.(type) {
		case *syntax.Function, *syntax.ClassLike:
			return false
		}
	}
	return true
}

// leavesFlow reports whether the block's last statement is return, throw,
// break, continue or an expression statement starting with exit/die.
func leavesFlow(ctx *analysis.Context, b *syntax.Block) bool {
	if len(b.Stmts) == 0 {
		return false
	}
	switch last := b.Stmts[len(b.Stmts)-1].(type) {
	case *syntax.Return, *syntax.Break, *syntax.Continue:
		return true
	case *syntax.ExprStmt:
		if _, ok := last.Expr.(*syntax.Throw); ok {
			return true
		}
		t, ok := astquery.NextSignificant(ctx.File, last.Span().Start)
		return ok && t.Kind == syntax.TExit && t.Start == last.Span().Start
	}
	return false
}
