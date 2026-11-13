package index

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/php/syntax"
)

func TestResolveDocTypes(t *testing.T) {
	classes := map[string]*Class{
		"Choice": {FQN: "Choice", Kind: syntax.KindEnum, Props: map[string]*Property{"value": {Type: "string"}}, Consts: map[string]*ClassConst{
			"A": {Case: true, Value: "'a'"}, "B": {Case: true, Value: "2"}, "C": {Case: true}, "EXTRA": {Type: "int"},
		}},
		"Pure":      {FQN: "Pure", Kind: syntax.KindEnum, Consts: map[string]*ClassConst{"A": {Case: true}}},
		"Constants": {FQN: "Constants", Consts: map[string]*ClassConst{"NUMBER": {Type: "int"}, "OTHER": {}, "ARR": {Type: "array", Value: "[1,2]"}}},
	}
	lookup := func(name string) *Class { return classes[name] }
	for _, tc := range []struct{ text, want string }{
		{"key-of<Choice>", "string"},
		{"key-of<Constants>", "?unknown"},
		{"Choice::A", `\Choice`},
		{"Choice::*", `\Choice|int`},
		{"Constants::NUMBER", "int"},
		{"Constants::OTHER", "mixed"},
		{"value-of<Constants::ARR>", "mixed"},
		{"key-of<Constants::ARR>", "?unknown"},
		{"value-of<Choice>", "string"},
		{"value-of<Pure>", "mixed"},
		{"Choice::Missing", "?unknown"},
		{"Unknown::A", "?unknown"},
		{"Choice", "?unknown"},
	} {
		got := ResolveDocType(tc.text, nil, lookup)
		if got.DocString() != tc.want {
			t.Errorf("%s=%s want %s", tc.text, got.DocString(), tc.want)
		}
	}
	if got := ResolveDocType("Alias::A", func(string) string { return "Choice" }, lookup); got.DocString() != `\Choice` {
		t.Fatal(got)
	}
}

func TestExtractConstantDocTypes(t *testing.T) {
	fs := extract(t, "const.php", `<?php namespace Demo;
 /** @return State::Active|null */ function result() {}
 /** @return Constants::NUMBER */ function number() {}
 /** @return Missing::A */ function missing() {}
 class Before {function method(){new class{};}}
 enum State {case Active;}
 class Constants {const int NUMBER=1;function method(){}}
 `)
	if fs.Functions[0].DocReturn != `\Demo\State|null` || fs.Functions[1].DocReturn != "int" || fs.Functions[2].DocReturn != `\Demo\Missing::A` {
		t.Fatalf("%+v %+v %+v", fs.Functions[0], fs.Functions[1], fs.Functions[2])
	}
}

func TestExtractBackedEnumAndUnsupportedGenericDocs(t *testing.T) {
	fs := extract(t, "backed.php", `<?php namespace Demo;
 /** @return value-of<State> */ function value() {}
 /** @return key-of<array<Unknown>> */ function unsupported() {}
 const PREFIX = 'prefix';
 enum State: string {case Active = PREFIX;}
 `)
	if got := fs.Functions[0].DocReturn; got != "string" {
		t.Fatalf("backed value = %s", got)
	}
	if got := fs.Functions[1].DocReturn; got != "mixed" {
		t.Fatalf("unsupported key = %s", got)
	}
}

// Parameter and property PHPDoc types keep their integer ranges and
// non-empty strings beside the plain type; a promoted property takes those
// of its constructor parameter, or of its own @var.
func TestExtractDocValues(t *testing.T) {
	fs := extract(t, "values.php", `<?php namespace Demo;
 /**
  * @param positive-int $n
  * @param string $s
  * @param non-empty-string|null $code
  */
 function f($n, $s, $code) {}
 class Cart {
   /** @var int<1, 9> */
   public $digit;
   /** @var int */
   public $plain;
   /** @param negative-int $offset */
   public function __construct(public $offset, /** @var non-empty-string */ public $label) {}
 }
 `)
	ps := fs.Functions[0].Params
	if ps[0].DocType != "int" || ps[0].DocValues != "positive-int" || ps[1].DocValues != "" || ps[2].DocValues != "non-empty-string|null" {
		t.Fatalf("%+v", ps)
	}
	props := fs.Classes[0].Props
	for name, want := range map[string]string{"digit": "int<1, 9>", "plain": "", "offset": "negative-int", "label": "non-empty-string"} {
		if p := props[name]; p == nil || p.DocValues != want {
			t.Errorf("%s: %+v want %q", name, p, want)
		}
	}
}

func BenchmarkExtractConstantDocTypes(b *testing.B) {
	var src strings.Builder
	src.WriteString("<?php namespace Demo; enum State {case Active;} class Constants {const int NUMBER=1;}\n")
	for i := range 100 {
		fmt.Fprintf(&src, "/** @return State::Active|Constants::NUMBER|null */ function result%d() {}\n", i)
	}
	f := syntax.Parse("docs.php", []byte(src.String()), syntax.Options{})
	b.ReportAllocs()
	for b.Loop() {
		Extract(f)
	}
}
