package syntax

import phpversion "custos/internal/php/version"

// ParseBest parses src at opt.Version and, when that yields syntax errors,
// re-parses with the newest grammar (permissive) and keeps the result with
// fewer errors. The interpreter that runs the code, not the project's minimum
// version, decides what parses (a PHP 8.1-targeting project may contain 8.4
// test fixtures); parsing at the target version first keeps old code that
// uses later keywords (`fn`, `match`, `enum`) as identifiers working.
func ParseBest(path string, src []byte, opt Options) *File {
	f := Parse(path, src, opt)
	if len(f.Errors) == 0 || (opt.Version >= phpversion.Max && opt.Permissive) {
		return f
	}
	alt := opt
	alt.Version, alt.Permissive = phpversion.Max, true
	// The losing tree is discarded: its tokens go to the next parse.
	g := Parse(path, src, alt)
	if len(g.Errors) < len(f.Errors) {
		f.Release()
		g.Version = opt.Version
		return g
	}
	g.Release()
	return f
}
