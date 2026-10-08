package phpunit

import "custos/internal/analysis"

// detectedPHPUnitVersion infers the project's PHPUnit version from the
// assertion API that PHPUnit\Framework\Assert declares in the index (vendor
// symbols included). ok is false when PHPUnit is not indexed. Only the
// thresholds the rules care about are distinguished:
//   - assertMatchesRegularExpression (added in 9.1)      → 91
//   - assertIsInt without assertInternalType (9.0)      → 90
//   - assertIsInt (7.5 / 8.x)                           → 80
//   - neither                                           → 70
//
// When PHPUnit is not indexed (no vendor directory, or one installed
// without dev dependencies) the CLI passes the lowest version allowed by
// composer.json's phpunit/phpunit constraint as the internal option
// "@composerPHPUnit" (custos).
func detectedPHPUnitVersion(ctx *analysis.Context) (int, bool) {
	v := ctx.Memo("phpunit.detectedVersion", func() any {
		ix := ctx.Index()
		const assert = `PHPUnit\Framework\Assert`
		if ix.Class(assert, ctx.PHP) == nil {
			return ctx.Int("@composerPHPUnit")
		}
		has := func(m string) bool { return ix.FindMethod(assert, m, ctx.PHP) != nil }
		switch {
		case has("assertMatchesRegularExpression"):
			return 91
		case has("assertIsInt") && !has("assertInternalType"):
			return 90
		case has("assertIsInt"):
			return 80
		default:
			return 70
		}
	}).(int)
	return v, v != 0
}
