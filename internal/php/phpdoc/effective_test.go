package phpdoc

import (
	"reflect"
	"strings"
	"testing"
)

func TestEffectiveParams(t *testing.T) {
	d := Parse(`/**
 * @param int $x
 * @psalm-param string $x
 * @phpstan-param array{'first name': int, 'a}b': string} $x
 * @param bool $x
 * @phpstan-param $y The description
 * @psalm-param $y User the owner
 * @param $y int
 * @phpstan-param $empty
 * @param float
 * @param
 * @param bool $z
 * @param string $z
 * @phpstan-param string $z
 * @phpstan-param int $z
 * @phpstan-param $z
 * @return void
 */`)
	want := []Param{{Type: "array{'first name':int,'a}b':string}", Name: "x"}, {Type: "User", Name: "y"}, {Type: "int", Name: "z"}}
	if got := d.EffectiveParams(); !reflect.DeepEqual(got, want) {
		t.Fatalf("EffectiveParams() = %#v, want %#v", got, want)
	}
	if got := d.Params(); len(got) != 7 || got[0] != (Param{Type: "int", Name: "x"}) {
		t.Fatalf("legacy Params() changed: %#v", got)
	}
	if got := (&Doc{}).EffectiveParams(); got != nil {
		t.Fatalf("empty params = %#v", got)
	}
}

func TestEffectiveVars(t *testing.T) {
	d := Parse(`/**
 * @var int $x
 * @psalm-var string $x
 * @phpstan-var $x array{'first name': int}
 * @phpstan-var bool $x
 * @psalm-var float $x
 * @var string $y
 * @phpstan-var $y
 * @var $z User
 * @psalm-var $z bool
 * @var bool
 * @psalm-var float
 * @phpstan-var
 * @param string $other
 */`)
	want := []Param{{Type: "array{'first name':int}", Name: "x"}, {Type: "string", Name: "y"}, {Type: "bool", Name: "z"}, {Type: "float"}}
	if got := d.EffectiveVars(); !reflect.DeepEqual(got, want) {
		t.Fatalf("EffectiveVars() = %#v, want %#v", got, want)
	}
	for name, want := range map[string]string{"x": "array{'first name':int}", "y": "float", "z": "bool", "missing": "float", "": "array{'first name':int}"} {
		if got := d.EffectiveVarType(name); got != want {
			t.Errorf("EffectiveVarType(%q) = %q, want %q", name, got, want)
		}
	}
	if got := d.VarType("x"); got != "int" {
		t.Fatalf("legacy VarType changed: %q", got)
	}
	if got := Parse("/** @phpstan-var string $x */").EffectiveVarType("y"); got != "" {
		t.Fatalf("nonmatching variable type = %q", got)
	}
	if got := (&Doc{}).EffectiveVars(); got != nil {
		t.Fatalf("empty vars = %#v", got)
	}
}

func TestEffectiveReturnType(t *testing.T) {
	for _, tt := range []struct {
		name string
		tags []Tag
		want string
	}{
		{"empty", nil, ""},
		{"unrelated", []Tag{{"param", "int $x"}}, ""},
		{"ordinary", []Tag{{"return", "int"}}, "int"},
		{"priority", []Tag{{"return", "int"}, {"psalm-return", "string"}, {"phpstan-return", "bool"}, {"return", "float"}}, "bool"},
		{"reverse", []Tag{{"phpstan-return", "bool"}, {"psalm-return", "string"}, {"return", "int"}}, "bool"},
		{"same tier first", []Tag{{"phpstan-return", "int"}, {"phpstan-return", "string"}}, "int"},
		{"empty fallback", []Tag{{"phpstan-return", ""}, {"psalm-return", ""}, {"return", "int"}}, "int"},
		{"empty then valid", []Tag{{"phpstan-return", ""}, {"phpstan-return", "string"}}, "string"},
		{"prose fallback", []Tag{{"phpstan-return", "The item returned"}, {"return", "int"}}, "int"},
		{"single word", []Tag{{"phpstan-return", "The"}}, "The"},
		{"quoted shape", []Tag{{"psalm-return", `array{'a b': string, "c>d": bool} description`}}, `array{'a b':string,"c>d":bool}`},
		{"conditional", []Tag{{"phpstan-return", "($x is int ? string : bool)"}}, "($x is int ? string : bool)"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			d := &Doc{Tags: tt.tags}
			if got := d.EffectiveReturnType(); got != tt.want {
				t.Fatalf("EffectiveReturnType() = %q, want %q", got, tt.want)
			}
		})
	}
	d := &Doc{Tags: []Tag{{"phpstan-return", "string"}, {"return", "int"}}}
	if got := d.ReturnType(); got != "int" {
		t.Fatalf("legacy ReturnType changed: %q", got)
	}
}

func TestEffectiveAnnotationLongType(t *testing.T) {
	// Selection uses existing normalization and leaves parsing limits to types.
	typ := "array{'key': " + strings.Repeat("A", MaxAliasLen+1) + "}"
	d := &Doc{Tags: []Tag{{"phpstan-param", typ + " $p"}, {"phpstan-var", typ + " $v"}, {"phpstan-return", typ}}}
	typ, _ = SplitType(typ)
	if got := d.EffectiveParams(); len(got) != 1 || got[0].Type != typ {
		t.Fatal("parameter type was truncated")
	}
	if got := d.EffectiveVarType("v"); got != typ {
		t.Fatal("variable type was truncated")
	}
	if got := d.EffectiveReturnType(); got != typ {
		t.Fatal("return type was truncated")
	}
}

func BenchmarkEffectiveAnnotations(b *testing.B) {
	d := Parse(`/**
 * @param array $row
 * @psalm-param array{id: int} $row
 * @phpstan-param array{id: int, name: string} $row
 * @param $fallback string
 * @var object $item
 * @phpstan-var $item Product
 * @return array
 * @phpstan-return list<Product>
 */`)
	b.ReportAllocs()
	for b.Loop() {
		_ = d.EffectiveParams()
		_ = d.EffectiveVars()
		_ = d.EffectiveVarType("item")
		_ = d.EffectiveReturnType()
	}
}
