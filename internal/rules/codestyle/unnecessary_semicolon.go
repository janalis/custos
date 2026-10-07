package codestyle

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// unnecessarySemicolon reports empty statements and the terminator of a
// short echo tag right before `?>`.
type unnecessarySemicolon struct{}

func init() { register(unnecessarySemicolon{}) }

const unnecessarySemicolonMsg = "Stray semicolon; remove it."

func (unnecessarySemicolon) ID() string { return "UnnecessarySemicolon" }

func (unnecessarySemicolon) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KNop, syntax.KEcho}
}

func (r unnecessarySemicolon) Check(ctx *analysis.Context, n syntax.Node) {
	if strings.HasSuffix(ctx.File.Path, ".blade.php") { // E3
		return
	}
	switch n := n.(type) {
	case *syntax.Nop:
		s := n.Span()
		if s.Len() != 1 || ctx.Src[s.Start] != ';' {
			return // zero-width recovery node
		}
		if isBodyNop(n) { // E1
			return
		}
		r.report(ctx, s)
	case *syntax.Echo:
		if !n.Short || len(n.Exprs) == 0 {
			return
		}
		s := n.Span()
		if s.Len() == 0 || ctx.Src[s.End-1] != ';' {
			return // terminated by `?>`
		}
		if next, ok := util.NextSignificant(ctx.File, s.End); ok && next.Kind != syntax.TCloseTag {
			return // E2: more code in the same tag
		}
		r.report(ctx, syntax.Span{Start: s.End - 1, End: s.End})
	}
}

// isBodyNop reports whether the empty statement is the direct body of an
// if/elseif/else clause, a loop or a declare construct.
func isBodyNop(n *syntax.Nop) bool {
	var body syntax.Stmt
	switch p := n.Parent().(type) {
	case *syntax.If:
		body = p.Body
	case *syntax.ElseIf:
		body = p.Body
	case *syntax.Else:
		body = p.Body
	case *syntax.While:
		body = p.Body
	case *syntax.DoWhile:
		body = p.Body
	case *syntax.For:
		body = p.Body
	case *syntax.Foreach:
		body = p.Body
	default: // a declare body is never a `;` Nop: `declare(x=1);` has no body
		return false
	}
	return body == syntax.Stmt(n)
}

func (unnecessarySemicolon) report(ctx *analysis.Context, semi syntax.Span) {
	f := ctx.File
	ctx.Report(semi, unnecessarySemicolonMsg, analysis.Fix{
		Title: "Remove the semicolon",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: util.WithLeadingWhitespace(f, semi)}}
		},
	})
}
