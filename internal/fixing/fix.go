// Package fixing applies quick-fix text edits, iterating until no fix applies.
package fixing

import (
	"context"
	"sort"
	"sync"

	"github.com/janalis/custos/internal/diagnostic"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
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
// and ends the loop. Fixes that would make the source require a newer PHP
// version than both the target and the input are left out (see
// WithinVersion).
func FixSource(e Analyzer, path string, src []byte, opt Options) Result {
	return FixSourceContext(context.Background(), e, path, src, opt)
}

// FixSourceContext is FixSource stopping between iterations once ctx is
// done.
func FixSourceContext(ctx context.Context, e Analyzer, path string, src []byte, opt Options) Result {
	res := Result{Source: src}
	f := syntax.ParseBest(path, src, opt.Parse)
	// Fixes left out for needing a newer PHP version, so that the fixes
	// they overlapped get their turn.
	tooNew := map[fixKey]bool{}
	for res.Iterations < MaxIterations && ctx.Err() == nil {
		var edits []diagnostic.TextEdit
		var groups []fixGroup
		for _, fd := range e.Analyze(f) {
			if len(fd.Fixes) == 0 || (opt.Filter != nil && !opt.Filter(fd)) {
				continue
			}
			key := keyOf(fd, res.Source)
			if tooNew[key] {
				continue
			}
			// All edits of one fix must apply together: group them by
			// checking for overlap against already collected edits first.
			fe, ok := Edits(fd.Fixes[0])
			if !ok || !sound(fe, len(res.Source)) || overlapsAny(fe, edits) {
				continue
			}
			edits = append(edits, fe...)
			groups = append(groups, fixGroup{fe, key})
		}
		if len(edits) == 0 {
			break
		}
		out, n := Apply(res.Source, edits)
		res.Iterations++
		next := syntax.ParseBest(path, out, opt.Parse)
		if len(next.Errors) <= len(f.Errors) && tooNewFor(opt.Parse.Version, next, versionOf(f)) {
			// Rare: keep only the fixes that are compatible together.
			out, n, next = compatibleEdits(path, res.Source, f, groups, opt.Parse, tooNew)
			if n == 0 {
				continue // the next analysis skips the fixes left out
			}
		}
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

// fixKey tells a fix from another across iterations, where offsets move:
// its rule, title and the text of its finding.
type fixKey struct{ rule, title, text string }

func keyOf(fd diagnostic.Finding, src []byte) fixKey {
	return fixKey{fd.Rule, fd.Fixes[0].Title, string(src[fd.Span.Start:fd.Span.End])}
}

type fixGroup struct {
	edits []diagnostic.TextEdit
	key   fixKey
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

// WithinVersion reports whether after, the parsed result of fixing before,
// needs no newer PHP version than both the parse target and before itself:
// a fix must not introduce syntax the project's PHP version lacks, while
// code that already needs a newer version is not made worse. before is
// only parsed when after needs a newer version than the target.
func WithinVersion(path string, before []byte, after *syntax.File, opt syntax.Options) bool {
	return !tooNewFor(opt.Version, after, func() phpversion.Version {
		v, _ := syntax.RequiredVersion(syntax.ParseBest(path, before, opt))
		return v
	})
}

// tooNewFor reports whether after needs a newer PHP version than both
// target and had(), the version the unfixed source needs (only computed
// when after exceeds the target).
func tooNewFor(target phpversion.Version, after *syntax.File, had func() phpversion.Version) bool {
	if target == 0 {
		target = phpversion.Default
	}
	if target >= phpversion.Max {
		return false
	}
	need, _ := syntax.RequiredVersion(after)
	return need > target && need > had()
}

// versionOf returns RequiredVersion(f), computed once on first call.
func versionOf(f *syntax.File) func() phpversion.Version {
	return sync.OnceValue(func() phpversion.Version {
		v, _ := syntax.RequiredVersion(f)
		return v
	})
}

// compatibleEdits applies to src, in order, each edit group that keeps the
// result within the target version and adds no syntax errors, and returns
// that result, the number of edits applied and its tree. A group too new on
// its own is recorded in tooNew; one only too new combined with the groups
// before it (two edits can form a token together, such as `*` and `*`) is
// left for the next iteration.
func compatibleEdits(path string, src []byte, f *syntax.File, groups []fixGroup, opt syntax.Options, tooNew map[fixKey]bool) (out []byte, n int, next *syntax.File) {
	had := versionOf(f)
	var kept []diagnostic.TextEdit
	out, next = src, f
	for _, g := range groups {
		try := append(kept[:len(kept):len(kept)], g.edits...)
		o, k := Apply(src, try)
		t := syntax.ParseBest(path, o, opt)
		if len(t.Errors) > len(f.Errors) {
			continue
		}
		if tooNewFor(opt.Version, t, had) {
			if len(kept) == 0 || tooNewAlone(path, src, g.edits, opt, had) {
				tooNew[g.key] = true
			}
			continue
		}
		kept, out, n, next = try, o, k, t
	}
	return out, n, next
}

// tooNewAlone reports whether edits, applied alone to src, make it need a
// newer PHP version than both the target and had().
func tooNewAlone(path string, src []byte, edits []diagnostic.TextEdit, opt syntax.Options, had func() phpversion.Version) bool {
	o, _ := Apply(src, edits)
	return tooNewFor(opt.Version, syntax.ParseBest(path, o, opt), had)
}
