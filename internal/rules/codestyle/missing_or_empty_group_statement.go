package codestyle

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// missingOrEmptyGroupStatement reports control structures whose body is not
// a braced block, or (optionally) is an empty one.
type missingOrEmptyGroupStatement struct{}

func init() { register(missingOrEmptyGroupStatement{}) }

func (missingOrEmptyGroupStatement) ID() string { return "MissingOrEmptyGroupStatement" }

func (missingOrEmptyGroupStatement) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KIf, syntax.KElseIf, syntax.KElse, syntax.KForeach,
		syntax.KFor, syntax.KWhile, syntax.KDoWhile}
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
	kw, _ := util.TokenAfter(ctx.File, n.Span().Start)
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
		ctx.Report(kwSpan, "Use a braced block for the body of this construct.", analysis.Fix{
			Title: "Wrap the body in braces",
			Edits: func() []analysis.TextEdit {
				return []analysis.TextEdit{{Span: syntax.Span{Start: bs.Start, End: bs.Start}, NewText: "{} "}}
			},
		})
		return
	}
	ctx.Report(kwSpan, "Use a braced block for the body of this construct.", analysis.Fix{ // D1
		Title: "Wrap the body in braces",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: bs, NewText: "{\n" + text + "\n}"}}
		},
	})
}
