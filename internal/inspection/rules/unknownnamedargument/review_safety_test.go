package unknownnamedargument_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/unknownnamedargument"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{rule.New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, source string
		findings     int
	}{
		{"overloaded builtin", "<?php session_set_cookie_params(lifetime_or_options: 1, path: '/', secure: true);", 0},
		{"renamed builtin parameter", "<?php function f($d) { dba_fetch(key: 'k', dba: $d); ldap_sasl_bind(ldap: $d, dn: null, mech: 'x'); }", 0},
		{"overloaded builtin method", "<?php function f(IntlCalendar $c) { $c->set(year: 1, month: 0, dayOfMonth: 1); }", 0},
		{"unverified extension", "<?php function f(Imagick $i, SNMP $s) { $i->addImage(image: $i); $s->get(objectId: 'x'); snmp_set_enum_print(enable: true); }", 0},
		{"verified builtin", "<?php strlen(str: 'a'); json_encode(val: 1);", 2},
		{"verified builtin method", "<?php function f(DateTime $d) { $d->format(fmt: 'c'); }", 1},
		{"valid builtin names", "<?php in_array(needle: 1, haystack: [], strict: true); function f(DateTime $d) { $d->format(format: 'c'); }", 0},
		{"override renames", "<?php class A { function m($a) {} } class B extends A { function m($b) {} } function f(A $o) { $o->m(b: 1); }", 0},
		{"no override declares it", "<?php class A { function m($a) {} } class B extends A { function m($b) {} } function f(A $o) { $o->m(c: 1); }", 1},
		{"interface implementation", "<?php interface I { function m($a); } final class C implements I { function m($b) {} } function f(I $o) { $o->m(b: 1); }", 0},
		{"interface without implementation", "<?php interface I { function m($a); } function f(I $o) { $o->m(b: 1); }", 1},
		{"abstract through this", "<?php abstract class A { abstract function m($a); function go() { $this->m(b: 1); } } class B extends A { function m($b) {} }", 0},
		{"late static binding", "<?php class A { static function m($a) {} static function go() { static::m(b: 1); } } class B extends A { static function m($b) {} }", 0},
		{"self is exact", "<?php class A { static function m($a) {} static function go() { self::m(b: 1); } } class B extends A { static function m($b) {} }", 1},
		{"named class is exact", "<?php class A { static function m($a) {} } class B extends A { static function m($b) {} } A::m(b: 1);", 1},
		{"private method", "<?php class A { private function m($a) {} function go() { $this->m(b: 1); } } class B extends A { function m($b) {} }", 1},
		{"final method", "<?php class A { final function m($a) {} } class B extends A {} function f(A $o) { $o->m(b: 1); }", 1},
		{"variadic override", "<?php class A { function m($a) {} } class B extends A { function m(...$b) {} } function f(A $o) { $o->m(b: 1); }", 0},
		{"trait receiver", "<?php trait T { abstract function m($a); function go() { $this->m(b: 1); } }", 0},
		{"magic method tag", "<?php /** @method void m(int $a) */ class A { function __call($n, $a) {} } function f(A $o) { $o->m(b: 1); }", 0},
		{"union receiver", "<?php class A { function m($a) {} } class B { function m($b) {} } function f(A|B $o) { $o->m(b: 1); $o->m(c: 1); }", 1},
		{"conditional declarations", "<?php if (PHP_VERSION_ID > 80000) { function c($a) {} } else { function c($b) {} } c(b: 1); c(z: 1);", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("review.php", []byte(tc.source), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
