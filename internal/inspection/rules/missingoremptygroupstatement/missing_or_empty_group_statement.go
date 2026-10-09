package missingoremptygroupstatement

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// missingOrEmptyGroupStatement reports control structures whose body is not
// a braced block, or (optionally) is an empty one.
type missingOrEmptyGroupStatement struct{}

func (missingOrEmptyGroupStatement) ID() string { return "MissingOrEmptyGroupStatement" }
func (missingOrEmptyGroupStatement) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{
		syntax.KIf, syntax.KElseIf, syntax.KElse, syntax.KForeach,
		syntax.KFor, syntax.KWhile, syntax.KDoWhile,
	}
}

func (missingOrEmptyGroupStatement) Check(ctx *analysis.Context, n syntax.Node) {
	if strings.HasSuffix(ctx.File.Path, ".blade.php") { // E2
		return
	}
	var body syntax.Stmt
	switch n := n.(type) {
	case *syntax.If:
		body = n.Body
	case *syntax.ElseIf:
		body = n.Body
	case *syntax.Else:
		if _, ok := n.Body.(*syntax.If); ok { // E1
			return
		}
		body = n.Body
	case *syntax.Foreach:
		body = n.Body
	case *syntax.For:
		body = n.Body
	case *syntax.While:
		body = n.Body
	case *syntax.DoWhile:
		body = n.Body
	}
	// The parser always gives these constructs a body (a Nop on error) and
	// starts their span at the keyword token.
	kw, _ := astquery.TokenAfter(ctx.File, n.Span().Start)
	kwSpan := syntax.Span{Start: kw.Start, End: kw.End}
	if b, ok := body.(*syntax.Block); ok {
		if len(b.Stmts) == 0 && ctx.Bool("REPORT_EMPTY_BODY") { // D2
			ctx.Report(kwSpan, "This construct has an empty body block.")
		}
		return // E4: alternative syntax never gets D1
	}
	bs := body.Span()
	if bs.Len() == 0 {
		return // recovery node
	}
	text := ctx.Text(body)
	if strings.HasPrefix(text, "?>") {
		// `if ($a) ?>html`: the close tag is the (empty) body; an empty block
		// before it keeps the inline HTML outside the construct.
		ctx.Report(kwSpan, "Use a braced block for the body of this construct.", diagnostic.Fix{
			Title: "Wrap the body in braces",
			Edits: func() []diagnostic.TextEdit {
				return []diagnostic.TextEdit{{Span: syntax.Span{Start: bs.Start, End: bs.Start}, NewText: "{} "}}
			},
		})
		return
	}
	ctx.Report(kwSpan, "Use a braced block for the body of this construct.", diagnostic.Fix{ // D1
		Title: "Wrap the body in braces",
		Edits: func() []diagnostic.TextEdit {
			// Indented under the construct's line (one level deeper).
			// The brace joins the header line: whitespace before the body
			// (a line break for `if ($a)\n    stmt;`) is replaced.
			indent := astquery.LineIndent(ctx.Src, kwSpan.Start)
			// custos: a comment between the header and the body (`for (…)
			// // note`) would swallow a brace placed after it; open the
			// block right after the header instead and keep the body
			// where it is.
			if at, ok := braceBeforeComment(ctx.File, bs.Start); ok {
				return []diagnostic.TextEdit{
					{Span: syntax.Span{Start: at, End: at}, NewText: " {"},
					{Span: bs, NewText: text + "\n" + indent + "}"},
				}
			}
			span, open := bs, "{\n"
			if ws, ok := astquery.TokenBefore(ctx.File, bs.Start); ok && ws.Kind == syntax.TWhitespace {
				span.Start, open = ws.Start, " {\n"
			}
			// custos: a statement ended by a close tag (`if ($h) echo "x" ?>`)
			// gets a `;` and the brace before the tag, not after it in the
			// HTML.
			if i := strings.LastIndex(text, "?>"); i > 0 && strings.TrimSpace(text[i+2:]) == "" {
				stmt := strings.TrimRight(text[:i], " \t\r\n")
				return []diagnostic.TextEdit{{Span: span, NewText: open + indent + "    " + stmt + ";\n" + indent + "} " + text[i:]}}
			}
			return []diagnostic.TextEdit{{Span: span, NewText: open + indent + "    " + text + "\n" + indent + "}"}}
		},
	})
}

// braceBeforeComment returns the end of the last significant token before
// off when a line comment (`//`, `#`) sits between it and off.
func braceBeforeComment(f *syntax.File, off uint32) (uint32, bool) {
	comment := false
	// The file's first token (inline HTML or the open tag) is significant,
	// so the walk always stops.
	for i := astquery.TokenIndex(f, off) - 1; ; i-- {
		t := f.Tokens[i]
		switch t.Kind {
		case syntax.TWhitespace, syntax.TDocComment:
		case syntax.TComment:
			comment = comment || f.Src[t.Start] != '/' || f.Src[t.Start+1] == '/'
		default:
			return f.Tokens[i].End, comment
		}
	}
}
