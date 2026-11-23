package syntax

import (
	"cmp"
	"slices"

	phpversion "github.com/janalis/custos/internal/php/version"
)

// ParseBest parses src at opt.Version and, when that yields syntax errors,
// re-parses with the newest grammar (permissive) and keeps the result with
// fewer errors. The interpreter that runs the code, not the project's minimum
// version, decides what parses (a PHP 8.1-targeting project may contain 8.4
// test fixtures); parsing at the target version first keeps old code that
// uses later keywords (`fn`, `match`, `enum`) as identifiers working.
func ParseBest(path string, src []byte, opt Options) *File {
	f := Parse(path, src, opt)
	// Compile errors (`f(a: 1, 2)`) leave the tree whole: they are the
	// target version's to report, and no newer grammar fixes them.
	if !f.Recovered() || (opt.Version >= phpversion.Max && opt.Permissive) {
		return f
	}
	alt := opt
	alt.Version, alt.Permissive = phpversion.Max, true
	if g := Parse(path, src, alt); g.RecoveryErrorCount() < f.RecoveryErrorCount() {
		g.Version = opt.Version
		g.Errors = keepCompileErrors(g.Errors, f.Errors)
		return g
	}
	return f
}

// keepCompileErrors adds to errs the compile errors of the target version's
// parse (from) that the newest grammar does not raise: some apply only below
// a version (`__debugInfo` on an enum before PHP 8.6). The result stays in
// source order.
func keepCompileErrors(errs, from []Error) []Error {
	n := len(errs)
	for _, e := range from {
		if e.Compile && !slices.ContainsFunc(errs[:n], func(g Error) bool { return g.Span == e.Span && g.Compile }) {
			errs = append(errs, e)
		}
	}
	if len(errs) > n {
		slices.SortStableFunc(errs, func(a, b Error) int { return cmp.Compare(a.Span.Start, b.Span.Start) })
	}
	return errs
}
