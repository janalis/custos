package unsetconstructscanbemerged

import (
	"bytes"
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// unsetConstructsCanBeMerged reports an unset() statement directly following
// another unset() statement in the same statement list.
type unsetConstructsCanBeMerged struct{}

const unsetConstructsCanBeMergedMsg = "Consecutive unset() calls; merge them into one."

// unsetMaxLine is the width past which a merge of statements written on
// several lines keeps one line per statement (F1a).
const unsetMaxLine = 120

func (unsetConstructsCanBeMerged) ID() string { return "UnsetConstructsCanBeMerged" }
func (unsetConstructsCanBeMerged) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KUnset}
}

func (unsetConstructsCanBeMerged) Check(ctx *analysis.Context, n syntax.Node) {
	cur := n.(*syntax.Unset)
	if cur.Span().Len() == 0 || len(cur.Vars) == 0 {
		return
	}
	list, i, ok := astquery.StmtList(ctx.File, cur)
	if !ok {
		return
	}
	if i <= 0 {
		return
	}
	prev, isUnset := list[i-1].(*syntax.Unset) // D1, D2
	if !isUnset || len(prev.Vars) == 0 {
		return
	}
	// The first statement of the run receives the merged arguments.
	firstIdx := i - 1
	for firstIdx > 0 {
		p, ok := list[firstIdx-1].(*syntax.Unset)
		if !ok || len(p.Vars) == 0 {
			break
		}
		firstIdx--
	}
	// End of the run (exclusive).
	end := i + 1
	for end < len(list) {
		if u, ok := list[end].(*syntax.Unset); !ok || len(u.Vars) == 0 {
			break
		}
		end++
	}
	f := ctx.File
	// F3: the merge regenerates one statement and deletes the others, so it
	// is offered only when every statement of the run ends with `;` (a
	// statement closed by `?>` owns the tag), holds no comment, and none
	// but the last has a comment after it on its line: that comment would
	// be appended to the merged line, away from the argument it describes.
	// Nor does it cross a blank line followed by a comment: that comment
	// opens a paragraph of its own, whose statements the merge moves away.
	for k, st := range list[firstIdx:end] {
		sp := st.Span()
		trail, paragraph := sp, false
		if next := firstIdx + k + 1; next < end {
			gap := syntax.Span{Start: sp.End, End: list[next].Span().Start}
			paragraph = hasBlankLine(f.Src[gap.Start:gap.End]) && astquery.HasComment(f, gap)
			trail = gap
			if nl := bytes.IndexByte(f.Src[sp.End:trail.End], '\n'); nl >= 0 {
				trail.End = sp.End + uint32(nl)
			}
		}
		if paragraph || f.Src[sp.End-1] != ';' || astquery.HasComment(f, sp) || astquery.HasComment(f, trail) {
			ctx.Report(cur.Span(), unsetConstructsCanBeMergedMsg)
			return
		}
	}
	// Merged statements: just this one, or — for the second statement of a
	// run — the whole run, so that applying every fix of the run at once
	// (single pass, non-overlapping edits) accumulates everything (F2).
	merged := list[i : i+1]
	if firstIdx == i-1 {
		merged = list[i:end]
	}
	first := list[firstIdx].(*syntax.Unset)
	ctx.Report(cur.Span(), unsetConstructsCanBeMergedMsg, diagnostic.Fix{
		Title: "Merge the unset() calls",
		Edits: func() []diagnostic.TextEdit {
			var args, groups []string
			for _, st := range append([]syntax.Stmt{first}, merged...) {
				var group []string
				for _, v := range st.(*syntax.Unset).Vars {
					group = append(group, ctx.Text(v))
				}
				args = append(args, group...)
				groups = append(groups, strings.Join(group, ", "))
			}
			// F1: regenerate the first statement (normalises `unset ( $x )`).
			text := "unset(" + strings.Join(args, ", ") + ");"
			indent := astquery.LineIndent(f.Src, first.Span().Start)
			last := merged[len(merged)-1].Span().End
			if len(indent)+len(text) > unsetMaxLine && bytes.IndexByte(f.Src[first.Span().End:last], '\n') >= 0 {
				// F1a: one line per merged statement rather than one long line.
				unit := astquery.IndentUnit(f.Src, indent)
				text = "unset(\n" + indent + unit + strings.Join(groups, ",\n"+indent+unit) + "\n" + indent + ");"
			}
			edits := []diagnostic.TextEdit{{Span: first.Span(), NewText: text}}
			for _, st := range merged {
				edits = append(edits, diagnostic.TextEdit{Span: astquery.WithLeadingWhitespace(f, st.Span())})
			}
			return edits
		},
	})
}

// hasBlankLine reports whether b holds a line made only of whitespace
// between two line breaks.
func hasBlankLine(b []byte) bool {
	lines := bytes.Split(b, []byte("\n"))
	if len(lines) < 3 {
		return false
	}
	for _, line := range lines[1 : len(lines)-1] {
		if len(bytes.TrimSpace(line)) == 0 {
			return true
		}
	}
	return false
}
