package phpunittests

import (
	"testing"

	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, source string
		findings     int
		fixed        string
	}{
		{"same-line tag kept", "<?php class T { /** @test @dataProvider rows */ function testA() {} function rows() {} }", 1, "<?php class T { /** @dataProvider rows */ function testA() {} function rows() {} }"},
		{"lone test tag", "<?php class T { /** @test */ function testA() {} }", 1, "<?php class T { /** */ function testA() {} }"},
		{"message constant", "<?php $t->assertSame($a, $b, true);", 0, ""},
		{"leading constant", "<?php $t->assertSame(true, $b, 'm');", 1, "<?php $t->assertTrue($b, 'm');"},
		{"syntax-only callable", "<?php $t->assertTrue(is_callable($f, true));", 0, ""},
		{"callable", "<?php $t->assertTrue(is_callable($f));", 1, "<?php $t->assertIsCallable($f);"},
		{"not a class name", "<?php $t->assertNotSame('not a class', get_class($o));", 0, ""},
		{"open class", "<?php class Animal {} $t->assertNotSame('Animal', get_class($o));", 1, ""},
		{"final class", "<?php final class Animal {} $t->assertNotSame('Animal', get_class($o));", 1, "<?php final class Animal {} $t->assertNotInstanceOf(\\Animal::class, $o);"},
		{"inverted unknown", "<?php $t->assertTrue(!$x);", 1, ""},
		{"inverted comparison", "<?php $t->assertTrue(!($x > 1));", 1, "<?php $t->assertNotTrue($x > 1);"},
		{"loose unknown", "<?php $t->assertTrue($a == $b);", 1, ""},
		{"loose integers", "<?php $t->assertTrue(1 == PHP_INT_SIZE);", 1, "<?php $t->assertEquals(1, PHP_INT_SIZE);"},
		{"identical unknown", "<?php $t->assertTrue($a === $b);", 1, ""},
		{"identical floats", "<?php $t->assertTrue(0.1 + 0.2 === 0.3);", 1, ""},
		{"identical literal", "<?php $t->assertTrue($a === 'x');", 1, "<?php $t->assertSame('x', $a);"},
		{"identical negative constant", "<?php $t->assertTrue($a === -PHP_INT_MAX);", 1, "<?php $t->assertSame(-PHP_INT_MAX, $a);"},
		{"identical class constant", "<?php class A { const B = 'b'; } $t->assertTrue($a === (A::B));", 1, "<?php class A { const B = 'b'; } $t->assertSame((A::B), $a);"},
		{"identical negation", "<?php $t->assertTrue(!$a === $b);", 1, "<?php $t->assertSame(!$a, $b);"},
		{"identical literals", "<?php $t->assertTrue(1 === 2);", 1, "<?php $t->assertSame(1, 2);"},
		{"empty offset", "<?php function f(string $s) { $t->assertTrue(empty($s[7])); }", 0, ""},
		{"empty global", "<?php $g = 'a'; $t->assertTrue(empty($g));", 0, ""},
		{"empty captured", "<?php function f() { $l = []; return function () use ($l) { $this->assertTrue(empty($l)); }; }", 0, ""},
		{"empty parameter", "<?php function f(string $s) { $t->assertTrue(empty($s)); }", 1, "<?php function f(string $s) { $t->assertEmpty($s); }"},
		{"count coerced", "<?php declare(strict_types=0); $t->assertSame($n, count($l));", 1, "<?php declare(strict_types=0); $t->assertCount($n, $l);"},
		{"count strict", "<?php declare(strict_types=1); $t->assertSame($n, count($l));", 1, ""},
		{"count string", "<?php $t->assertEquals('2', count($l));", 1, ""},
		{"instanceof object", "<?php $b = new stdClass(); $t->assertTrue($a instanceof $b);", 1, ""},
		{"file against integer", "<?php $t->assertEquals(file_get_contents($f), 42);", 1, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte(tc.source)
			opt := syntax.Options{Version: phpversion.PHP85}
			f := syntax.ParseBest("review.php", src, opt)
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			got := e.Analyze(f)
			if len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
			res := fixing.FixSource(e, "review.php", src, fixing.Options{Parse: opt, SinglePass: true})
			want := tc.source
			if tc.fixed != "" {
				want = tc.fixed
			}
			if string(res.Source) != want {
				t.Fatalf("fixed source: got %s, want %s", res.Source, want)
			}
			if len(syntax.ParseBest("review.php", res.Source, opt).Errors) != 0 {
				t.Fatal("fix introduced syntax errors")
			}
		})
	}
}
