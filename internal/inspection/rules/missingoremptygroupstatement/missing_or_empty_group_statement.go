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

// Semantic marks the rule as needing types: an empty foreach over an
// iterator runs it for its effects (E5).
func (missingOrEmptyGroupStatement) Semantic() {}

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
		if fe, ok := n.(*syntax.Foreach); ok && (iterates(ctx, fe.Expr) || leavesLast(ctx.File, fe)) { // E5, E6
			return
		}
		if len(b.Stmts) == 0 && ctx.Bool("REPORT_EMPTY_BODY") && !skipBranch(ctx.File, n, b) { // D2, E7
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
				// custos: after a line comment (`for (…) # note ?>`) the
				// block would be commented out; it goes before the comment.
				if at, ok := braceBeforeComment(ctx.File, bs.Start); ok {
					return []diagnostic.TextEdit{{Span: syntax.Span{Start: at, End: at}, NewText: " {}"}}
				}
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
			// custos: a statement ended by a close tag (`if ($h) echo "x" ?>`)
			// gets a `;` and the brace before the tag, not after it in the
			// HTML; both go before any comment that precedes the tag.
			stmt, tail, end := text, "", bs.End
			if at, ok := closeTagSplit(ctx.File, bs); ok {
				stmt, tail, end = text[:at-bs.Start]+";", text[at-bs.Start:], at
				if strings.HasPrefix(tail, "?>") {
					tail = " " + tail
				}
			}
			// custos: a one-line template island (`<?php if ($a) print "x"
			// ?>` inside an attribute) keeps its block on that line.
			inline := !strings.Contains(string(ctx.Src[kwSpan.Start:end]), "\n") && (tail != "" || closesAfter(ctx.File, bs.End))
			closing := "\n" + indent + "}" + tail
			if inline {
				closing = " }" + tail
			}
			// custos: a comment between the header and the body (`for (…)
			// // note`) would swallow a brace placed after it; open the
			// block right after the header instead and keep the body
			// where it is.
			if at, ok := braceBeforeComment(ctx.File, bs.Start); ok {
				return []diagnostic.TextEdit{
					{Span: syntax.Span{Start: at, End: at}, NewText: " {"},
					{Span: bs, NewText: stmt + closing},
				}
			}
			span, open, lead := bs, "{\n", indent+"    "
			if inline {
				open, lead = "{ ", ""
			}
			if ws, ok := astquery.TokenBefore(ctx.File, bs.Start); ok && ws.Kind == syntax.TWhitespace {
				span.Start, open = ws.Start, " "+open
			}
			return []diagnostic.TextEdit{{Span: span, NewText: open + lead + stmt + closing}}
		},
	})
}

// iterates reports whether e is known to be an iterator rather than an
// array: `iterable`, \Traversable or a subtype of it (\Generator included).
// Looping over it without a body drains it for its effects.
func iterates(ctx *analysis.Context, e syntax.Expr) bool {
	for _, a := range ctx.TypeOf(e).Atoms() {
		if strings.EqualFold(a, "iterable") {
			return true
		}
		if strings.HasPrefix(a, `\`) && !strings.HasSuffix(a, "]") && ctx.Index().IsSubtype(a, `\Traversable`, ctx.PHP) {
			return true
		}
	}
	return false
}

// skipBranch reports whether an empty if or elseif body is a commented
// branch that keeps the rest of its chain from running (E7).
func skipBranch(f *syntax.File, n syntax.Node, b *syntax.Block) bool {
	var chain *syntax.If
	switch n := n.(type) {
	case *syntax.If:
		chain = n
	case *syntax.ElseIf:
		chain, _ = n.Parent().(*syntax.If)
	}
	if chain == nil || !astquery.HasComment(f, b.Span()) {
		return false
	}
	return chain.Else != nil || len(chain.ElseIfs) > 0 && syntax.Node(chain.ElseIfs[len(chain.ElseIfs)-1]) != n
}

// leavesLast reports whether an empty foreach is run for the last key or
// value it leaves: a target other than a plain variable (`$stub->class`,
// `$row['k']`), or a variable mentioned after the loop (E6).
func leavesLast(f *syntax.File, fe *syntax.Foreach) bool {
	names := map[string]bool{}
	for _, t := range []syntax.Expr{fe.Key, fe.Value} {
		switch t := t.(type) {
		case nil:
		case *syntax.Variable:
			names[t.Name] = true
		default:
			return true
		}
	}
	var scope syntax.Node = f
	if fn := syntax.EnclosingFuncLike(fe); fn != nil {
		scope = fn
	}
	found := false
	syntax.Inspect(scope, func(x syntax.Node) bool {
		if v, ok := x.(*syntax.Variable); ok && v.Span().Start >= fe.Span().End && (v.NameExpr != nil || names[v.Name]) {
			found = true
		}
		return !found
	})
	return found
}

// closesAfter reports whether a close tag follows pos on the same line.
func closesAfter(f *syntax.File, pos uint32) bool {
	i := astquery.TokenIndex(f, pos)
	for i < len(f.Tokens) && f.Tokens[i].Kind == syntax.TWhitespace && !strings.Contains(string(f.Src[f.Tokens[i].Start:f.Tokens[i].End]), "\n") {
		i++
	}
	return i < len(f.Tokens) && f.Tokens[i].Kind == syntax.TCloseTag
}

// closeTagSplit returns the end of the body's last significant token when a
// close tag, not a `;`, ends the body.
func closeTagSplit(f *syntax.File, body syntax.Span) (uint32, bool) {
	// The body is not empty and does not start with the tag, so the walk
	// stops on one of its own tokens.
	i := astquery.TokenIndex(f, body.End) - 1
	if f.Tokens[i].Kind != syntax.TCloseTag {
		return 0, false
	}
	for i--; f.Tokens[i].Kind == syntax.TWhitespace || f.Tokens[i].Kind == syntax.TComment || f.Tokens[i].Kind == syntax.TDocComment; i-- {
	}
	return f.Tokens[i].End, true
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
