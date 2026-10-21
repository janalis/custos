// Package fixing applies quick-fix text edits, iterating until no fix applies.
package fixing

import (
	"context"
	"sort"

	"custos/internal/diagnostic"
	"custos/internal/php/syntax"
)

// Analyzer supplies findings for a parsed source file.
type Analyzer interface {
	Analyze(*syntax.File) []diagnostic.Finding
}

// MaxIterations bounds the re-analyse/re-fix loop.
const MaxIterations = 10

// Apply applies edits to src. Edits overlapping an earlier (by position,
// then by input order) edit are dropped; the number applied is returned.
func Apply(src []byte, edits []diagnostic.TextEdit) ([]byte, int) {
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
	Filter func(diagnostic.Finding) bool
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
// changes or MaxIterations is reached. A fix applies whole or not at all:
// one whose edits leave the source or overlap each other is skipped, and an
// iteration whose result has more syntax errors than its input is undone
// and ends the loop.
func FixSource(e Analyzer, path string, src []byte, opt Options) Result {
	return FixSourceContext(context.Background(), e, path, src, opt)
}

// FixSourceContext is FixSource stopping between iterations once ctx is
// done.
func FixSourceContext(ctx context.Context, e Analyzer, path string, src []byte, opt Options) Result {
	res := Result{Source: src}
	f := syntax.ParseBest(path, src, opt.Parse)
	for res.Iterations < MaxIterations && ctx.Err() == nil {
		var edits []diagnostic.TextEdit
		for _, fd := range e.Analyze(f) {
			if len(fd.Fixes) == 0 || (opt.Filter != nil && !opt.Filter(fd)) {
				continue
			}
			// All edits of one fix must apply together: group them by
			// checking for overlap against already collected edits first.
			fe, ok := Edits(fd.Fixes[0])
			if !ok || !sound(fe, len(res.Source)) || overlapsAny(fe, edits) {
				continue
			}
			edits = append(edits, fe...)
		}
		if len(edits) == 0 {
			break
		}
		out, n := Apply(res.Source, edits)
		res.Iterations++
		next := syntax.ParseBest(path, out, opt.Parse)
		if len(next.Errors) > len(f.Errors) {
			break
		}
		res.Applied += n
		res.Source, f = out, next
		if opt.SinglePass {
			break
		}
	}
	return res
}

// sound reports that the edits of one fix lie within a source of size bytes
// and do not overlap each other, so Apply keeps all of them. Unlike
// overlapsAny, edits that only touch are fine: one fix orders its own.
func sound(edits []diagnostic.TextEdit, size int) bool {
	for i, x := range edits {
		if x.Span.End < x.Span.Start || x.Span.End > uint32(size) {
			return false
		}
		for _, y := range edits[i+1:] {
			if x.Span.Start < y.Span.End && y.Span.Start < x.Span.End {
				return false
			}
		}
	}
	return true
}

// Edits computes the edits of a fix. Fixes compute them lazily, outside the
// engine's crash isolation: ok is false when the computation panics, and
// the fix is then left out instead of failing the whole request or run.
func Edits(f diagnostic.Fix) (edits []diagnostic.TextEdit, ok bool) {
	defer func() {
		if recover() != nil {
			edits, ok = nil, false
		}
	}()
	return f.Edits(), true
}

func overlapsAny(a, b []diagnostic.TextEdit) bool {
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
