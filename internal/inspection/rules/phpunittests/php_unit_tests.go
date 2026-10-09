package phpunittests

import (
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/phpunit"
	"custos/internal/php/syntax"
)

// phpUnitTests validates PHPUnit docblock tags (@covers, @depends,
// @dataProvider, @test) and promotes dedicated assertion/mocking APIs.
// Part A lives in php_unit_tests_tags.go, Part B in php_unit_tests_asserts.go.
type phpUnitTests struct{}

func (phpUnitTests) ID() string { return "PhpUnitTests" }
func (phpUnitTests) Semantic()  {}
func (phpUnitTests) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KMethod, syntax.KMethodCall, syntax.KStaticCall}
}

func (phpUnitTests) Check(ctx *analysis.Context, n syntax.Node) {
	if m, ok := n.(*syntax.Method); ok {
		putCheckTags(ctx, m)
		return
	}
	c, ok := phpunit.AsPuCall(n)
	if !ok {
		return
	}
	putCheckCall(ctx, c)
}

// putVersion returns the configured PHPUnit version as a number (80 for
// PHPUNIT80). When the option is not configured, the version is inferred
// from the indexed PHPUnit (detectedPHPUnitVersion). The EA harness strips
// enum prefixes (`PhpUnitVersion.PHPUNIT75`) before options reach the rule.
func putVersion(ctx *analysis.Context) int {
	if !ctx.OptionSet("PHP_UNIT_VERSION") {
		if v, ok := phpunit.DetectedPHPUnitVersion(ctx); ok {
			return v
		}
	}
	v := strings.ToUpper(ctx.String("PHP_UNIT_VERSION"))
	if strings.HasPrefix(v, "PHPUNIT") {
		if n, err := strconv.Atoi(v[len("PHPUNIT"):]); err == nil {
			return n
		}
	}
	return 80
}
