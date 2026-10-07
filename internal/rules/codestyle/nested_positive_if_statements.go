package codestyle

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// nestedPositiveIfStatements reports an if that is the only statement of a
// parent if (mergeable with &&) or of an else block (becomes else if).
type nestedPositiveIfStatements struct{}

func init() { register(nestedPositiveIfStatements{}) }

func (nestedPositiveIfStatements) ID() string { return "NestedPositiveIfStatements" }

func (nestedPositiveIfStatements) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KIf} }

const nestedPositiveIfMsg = "Merge this if statement into its parent construct."

func (nestedPositiveIfStatements) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.If)
	block, ok := c.Parent().(*syntax.Block)
	if !ok || block.Alt || len(block.Stmts) != 1 || c.Cond == nil { // D3, D7, E2, E6
		return
	}
	kw, ok := util.TokenAfter(ctx.File, c.Span().Start)
	if !ok || kw.Kind != syntax.TIf {
		return
	}
	kwSpan := syntax.Span{Start: kw.Start, End: kw.End}
	f := ctx.File
	switch p := block.Parent().(type) {
	case *syntax.If: // Case A
		if p.Body != syntax.Stmt(block) || p.Cond == nil { // D1
			return
		}
		if npiIsOr(p.Cond) || npiIsOr(c.Cond) { // D2, E3
			return
		}
		if len(p.ElseIfs) > 0 || len(c.ElseIfs) > 0 { // D4, E5
			return
		}
		if !npiElseCompatible(f, p.Else, c.Else) { // D5, E4
			return
		}
		ctx.Report(kwSpan, nestedPositiveIfMsg, analysis.Fix{
			Title: "Merge into the parent if",
			Edits: func() []analysis.TextEdit {
				pc := string(f.Src[p.Cond.Span().Start:p.Cond.Span().End])
				if npiWrapParent(p.Cond) {
					pc = "(" + pc + ")"
				}
				cc := string(f.Src[c.Cond.Span().Start:c.Cond.Span().End])
				body := npiWithComments(f, block, c, c.Body)
				return []analysis.TextEdit{
					{Span: p.Cond.Span(), NewText: pc + " && " + cc},
					{Span: block.Span(), NewText: body},
				}
			},
		})
	case *syntax.Else: // Case B
		if p.Body != syntax.Stmt(block) { // D6
			return
		}
		ctx.Report(kwSpan, nestedPositiveIfMsg, analysis.Fix{
			Title: "Turn into else if",
			Edits: func() []analysis.TextEdit {
				return []analysis.TextEdit{{Span: block.Span(), NewText: npiWithComments(f, block, c, c)}}
			},
		})
	}
}

func npiIsOr(e syntax.Expr) bool {
	b, ok := e.(*syntax.Binary)
	return ok && (b.Op.Kind == syntax.TBooleanOr || b.Op.Kind == syntax.TOr)
}

func npiWrapParent(e syntax.Expr) bool {
	switch e := e.(type) {
	case *syntax.Assign, *syntax.Ternary, *syntax.Instanceof:
		return true
	case *syntax.Binary:
		return e.Op.Kind != syntax.TBooleanAnd
	}
	return false
}

// npiElseCompatible implements D5.
func npiElseCompatible(f *syntax.File, pe, ce *syntax.Else) bool {
	if ce == nil {
		return pe == nil
	}
	if pe == nil {
		return false
	}
	pb, ok1 := pe.Body.(*syntax.Block)
	cb, ok2 := ce.Body.(*syntax.Block)
	if !ok1 || !ok2 || pb.Alt || cb.Alt || len(pb.Stmts) != len(cb.Stmts) {
		return false
	}
	return npiSameTokens(f, pb.Span(), cb.Span())
}

// npiSameTokens compares the significant token texts of two spans.
func npiSameTokens(f *syntax.File, a, b syntax.Span) bool {
	i, j := util.TokenIndex(f, a.Start), util.TokenIndex(f, b.Start)
	next := func(k int, s syntax.Span) (int, bool) {
		for ; k < len(f.Tokens) && f.Tokens[k].End <= s.End; k++ {
			switch f.Tokens[k].Kind {
			case syntax.TWhitespace, syntax.TComment, syntax.TDocComment:
				continue
			}
			return k, true
		}
		return k, false
	}
	for {
		var okA, okB bool
		i, okA = next(i, a)
		j, okB = next(j, b)
		if !okA || !okB {
			return okA == okB
		}
		ta, tb := f.Tokens[i], f.Tokens[j]
		if string(f.Src[ta.Start:ta.End]) != string(f.Src[tb.Start:tb.End]) {
			return false
		}
		i++
		j++
	}
}

// npiWithComments returns the text of n where comments lying in block before
// c are inserted right after the opening brace of c's body (when braced).
func npiWithComments(f *syntax.File, block *syntax.Block, c *syntax.If, n syntax.Node) string {
	ns := n.Span()
	text := string(f.Src[ns.Start:ns.End])
	body, ok := c.Body.(*syntax.Block)
	if !ok || body.Alt || body.Span().Len() == 0 || f.Src[body.Span().Start] != '{' {
		return text
	}
	var comments []string
	for k := util.TokenIndex(f, block.Span().Start); k < len(f.Tokens); k++ {
		t := f.Tokens[k]
		if t.End > c.Span().Start {
			break
		}
		if t.Kind == syntax.TComment || t.Kind == syntax.TDocComment {
			comments = append(comments, string(f.Src[t.Start:t.End]))
		}
	}
	if len(comments) == 0 {
		return text
	}
	at := body.Span().Start + 1 - ns.Start
	indent := naIndent(f.Src, c.Span().Start)
	var b strings.Builder
	b.WriteString(text[:at])
	for _, cm := range comments {
		b.WriteString("\n" + indent + cm)
	}
	rest := text[at:]
	if t := strings.TrimLeft(rest, " \t"); !strings.HasPrefix(t, "\n") && !strings.HasPrefix(t, "\r") {
		b.WriteString("\n" + indent)
	}
	b.WriteString(rest)
	return b.String()
}
