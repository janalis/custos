package unsetconstructscanbemerged

import (
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
	// Merged statements: just this one, or — for the second statement of a
	// run — the whole run, so that applying every fix of the run at once
	// (single pass, non-overlapping edits) accumulates everything (F2).
	merged := list[i : i+1]
	if firstIdx == i-1 {
		j := i + 1
		for j < len(list) {
			if u, ok := list[j].(*syntax.Unset); !ok || len(u.Vars) == 0 {
				break
			}
			j++
		}
		merged = list[i:j]
	}
	first := list[firstIdx].(*syntax.Unset)
	f := ctx.File
	ctx.Report(cur.Span(), unsetConstructsCanBeMergedMsg, diagnostic.Fix{
		Title: "Merge the unset() calls",
		Edits: func() []diagnostic.TextEdit {
			var args []string
			for _, st := range append([]syntax.Stmt{first}, merged...) {
				for _, v := range st.(*syntax.Unset).Vars {
					args = append(args, ctx.Text(v))
				}
			}
			// F1: regenerate the first statement (normalises `unset ( $x )`).
			edits := []diagnostic.TextEdit{{Span: first.Span(), NewText: "unset(" + strings.Join(args, ", ") + ");"}}
			for _, st := range merged {
				edits = append(edits, diagnostic.TextEdit{Span: astquery.WithLeadingWhitespace(f, st.Span())})
			}
			return edits
		},
	})
}
