package phpunit

import (
	"strconv"
	"strings"

	"custos/internal/analysis"
	"custos/internal/syntax"
)

// phpUnitTests validates PHPUnit docblock tags (@covers, @depends,
// @dataProvider, @test) and promotes dedicated assertion/mocking APIs.
// Part A lives in php_unit_tests_tags.go, Part B in php_unit_tests_asserts.go.
type phpUnitTests struct{}

func init() { register(phpUnitTests{}) }

func (phpUnitTests) ID() string { return "PhpUnitTests" }

func (phpUnitTests) Semantic() {}

func (phpUnitTests) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KMethod, syntax.KMethodCall, syntax.KStaticCall}
}

func (phpUnitTests) Check(ctx *analysis.Context, n syntax.Node) {
	if n.Span().Len() == 0 {
		return
	}
	if m, ok := n.(*syntax.Method); ok {
		putCheckTags(ctx, m)
		return
	}
	c, ok := asPuCall(n)
	if !ok {
		return
	}
	putCheckCall(ctx, c)
}

// putVersion returns the configured PHPUnit version as a number (80 for
// PHPUNIT80). When the option is not configured, the version is inferred
// from the indexed PHPUnit (detectedPHPUnitVersion). Values may carry an enum prefix (`PhpUnitVersion.PHPUNIT75`).
func putVersion(ctx *analysis.Context) int {
	if !ctx.OptionSet("PHP_UNIT_VERSION") {
		if v, ok := detectedPHPUnitVersion(ctx); ok {
			return v
		}
	}
	v := strings.ToUpper(ctx.String("PHP_UNIT_VERSION"))
	if i := strings.LastIndexByte(v, '.'); i >= 0 {
		v = v[i+1:]
	}
	if strings.HasPrefix(v, "PHPUNIT") {
		if n, err := strconv.Atoi(v[len("PHPUNIT"):]); err == nil {
			return n
		}
	}
	return 80
}
