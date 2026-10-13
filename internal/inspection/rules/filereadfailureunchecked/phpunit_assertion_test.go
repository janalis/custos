package filereadfailureunchecked

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// An assertion or expectation PHPUnit declares is the check of the value it
// is given, not a use of it (D1a); what the project declares stays a callee.
func TestPHPUnitAssertionsAreNoConsumption(t *testing.T) {
	const (
		value   = "file_get_contents($a)"
		library = "namespace PHPUnit\\Framework { abstract class Assert { public static function assertStringContainsString(string $needle, string $haystack): void {} } " +
			"abstract class TestCase extends Assert { public function expectOutputString(string $expected): void {} public function setName(string $name): void {} } " +
			"function assertStringContainsString(string $needle, string $haystack): void {} } "
		legacy = "namespace { abstract class PHPUnit_Framework_Assert { public static function assertContains(string $needle, string $haystack) {} } " +
			"abstract class PHPUnit_Framework_TestCase extends PHPUnit_Framework_Assert { public function expectOutputString(string $expected) {} } } "
		base = "namespace App { abstract class Base extends \\PHPUnit\\Framework\\TestCase { " +
			"protected function createRequest(string $body) {} protected static function assertReport(string $text) {} } } "
	)
	test := func(body string) string {
		return "namespace App { final class ReportTest extends Base { function test($a) { " + body + " } } }"
	}
	for _, tc := range []struct {
		name, source string
		findings     int
	}{
		{"assertion on $this", library + base + test("$this->assertStringContainsString('x', "+value+");"), 0},
		{"assertion in any case", library + base + test("$this->ASSERTStringContainsString('x', "+value+");"), 0},
		{"assertion through a variable", library + base + test("$v = "+value+"; $this->assertStringContainsString('x', $v);"), 0},
		{"assertion on self", library + base + test("self::assertStringContainsString('x', "+value+");"), 0},
		{"assertion on static", library + base + test("static::assertStringContainsString('x', "+value+");"), 0},
		{"assertion on parent", library + base + test("parent::assertStringContainsString('x', "+value+");"), 0},
		{"assertion on the class", library + base + test("\\PHPUnit\\Framework\\Assert::assertStringContainsString('x', "+value+");"), 0},
		{"assertion on a typed object", library + "namespace App { function f(\\PHPUnit\\Framework\\TestCase $t, $a) { $t->assertStringContainsString('x', " + value + "); } }", 0},
		{"assertion function", library + "namespace App { use function PHPUnit\\Framework\\assertStringContainsString; function f($a) { assertStringContainsString('x', " + value + "); } }", 0},
		{"expectation", library + base + test("$this->expectOutputString("+value+");"), 0},
		{"names before namespaces", legacy + "namespace { final class OldTest extends PHPUnit_Framework_TestCase { function test($a) { $this->assertContains('x', " + value + "); $this->expectOutputString(" + value + "); } } }", 0},
		{"another method of the test case", library + base + test("$this->setName("+value+");"), 1},
		{"helper of the project", library + base + test("$this->createRequest("+value+");"), 1},
		{"assertion of the project", library + base + test("self::assertReport("+value+");"), 1},
		{"assertion the project replaces", library + "namespace App { abstract class Base extends \\PHPUnit\\Framework\\TestCase { public static function assertStringContainsString(string $needle, string $haystack): void {} } } " +
			test("$this->assertStringContainsString('x', "+value+");"), 1},
		{"same names outside PHPUnit", "namespace Lib { abstract class Assert { public static function assertStringContainsString(string $needle, string $haystack): void {} } " +
			"function assertStringContainsString(string $needle, string $haystack): void {} } " +
			"namespace App { final class ReportTest extends \\Lib\\Assert { function test($a) { $this->assertStringContainsString('x', " + value + "); } } }", 1},
		{"same function outside PHPUnit", "namespace Lib { function assertStringContainsString(string $needle, string $haystack): void {} } " +
			"namespace App { function f($a) { \\Lib\\assertStringContainsString('x', " + value + "); } }", 1},
		{"PHPUnit declared twice", "namespace PHPUnit\\Framework { if ($c) { abstract class Assert { public static function assertStringContainsString(string $needle, string $haystack): void {} } } " +
			"else { abstract class Assert { public static function assertStringContainsString(string $needle, string $haystack, string $message = ''): void {} } } abstract class TestCase extends Assert {} } " +
			base + test("$this->assertStringContainsString('x', "+value+");"), 0},
		{"PHPUnit not declared", base + test("$this->assertStringContainsString('x', "+value+");"), 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.ParseBest("probe.php", []byte("<?php "+tc.source), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
