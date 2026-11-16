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
	if g := Parse(path, src, alt); len(g.Errors) < len(f.Errors) {
		g.Version = opt.Version
		return g
	}
	return f
}
