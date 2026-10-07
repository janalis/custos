// Package fix applies quick-fix text edits, iterating until no fix applies.
package fix

import (
	"sort"

	"custos/internal/analysis"
	"custos/internal/syntax"
)

// MaxIterations bounds the re-analyse/re-fix loop.
const MaxIterations = 10

// Apply applies edits to src. Edits overlapping an earlier (by position,
// then by input order) edit are dropped; the number applied is returned.
func Apply(src []byte, edits []analysis.TextEdit) ([]byte, int) {
	if len(edits) == 0 {
		return src, 0
	}
	idx := make([]int, len(edits))
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(a, b int) bool {
		ea, eb := edits[idx[a]], edits[idx[b]]
		if ea.Span.Start != eb.Span.Start {
			return ea.Span.Start < eb.Span.Start
		}
		return ea.Span.End < eb.Span.End
	})
	out := make([]byte, 0, len(src))
	pos := uint32(0)
	applied := 0
	for _, i := range idx {
		e := edits[i]
		if e.Span.Start < pos || e.Span.End > uint32(len(src)) || e.Span.End < e.Span.Start {
			continue // overlaps a previous edit or invalid
		}
		out = append(out, src[pos:e.Span.Start]...)
		out = append(out, e.NewText...)
		pos = e.Span.End
		applied++
	}
	out = append(out, src[pos:]...)
	return out, applied
}

// Options control FixSource.
type Options struct {
	Parse syntax.Options
	// Filter selects which findings to fix (nil = all fixable findings).
	Filter func(analysis.Finding) bool
	// SinglePass applies the fixes of the first analysis only (IDE
	// "apply all quick-fixes" semantics, used by conformance tests).
	SinglePass bool
}

// Result of fixing one source.
type Result struct {
	Source     []byte
	Applied    int // total edits applied over all iterations
	Iterations int
}

// FixSource repeatedly analyses src and applies the first fix of every
// fixable finding (one fix per finding, non-overlapping) until nothing
// changes or MaxIterations is reached.
func FixSource(e *analysis.Engine, path string, src []byte, opt Options) Result {
	res := Result{Source: src}
	for res.Iterations < MaxIterations {
		f := syntax.ParseBest(path, res.Source, opt.Parse)
		var edits []analysis.TextEdit
		for _, fd := range e.Analyze(f) {
			if len(fd.Fixes) == 0 || (opt.Filter != nil && !opt.Filter(fd)) {
				continue
			}
			// All edits of one fix must apply together: group them by
			// checking for overlap against already collected edits first.
			fe := fd.Fixes[0].Edits()
			if overlapsAny(fe, edits) {
				continue
			}
			edits = append(edits, fe...)
		}
		if len(edits) == 0 {
			break
		}
		out, n := Apply(res.Source, edits)
		res.Iterations++
		if n == 0 {
			break
		}
		res.Applied += n
		res.Source = out
		if opt.SinglePass {
			break
		}
	}
	return res
}

func overlapsAny(a, b []analysis.TextEdit) bool {
	for _, x := range a {
		for _, y := range b {
			if x.Span.Start < y.Span.End && y.Span.Start < x.Span.End {
				return true
			}
			// An insertion touching another edit (two insertions at the same
			// point, or an insertion at either end of a replacement) is
			// ambiguous: `\` inserted before a call that another fix
			// rewrites would land before the rewritten text. Keep the first;
			// the next iteration re-analyses the result.
			if (x.Span.Start == x.Span.End && y.Span.Start <= x.Span.Start && x.Span.Start <= y.Span.End) ||
				(y.Span.Start == y.Span.End && x.Span.Start <= y.Span.Start && y.Span.Start <= x.Span.End) {
				return true
			}
		}
	}
	return false
}
