package semanticquery

import (
	"testing"

	"custos/internal/php/syntax"
)

func TestResolvedValueDiscoveryArgumentEffects(t *testing.T) {
	for _, tc := range []struct{ code, want string }{
		{`function mutate(&$value){$value='changed';} function example(){ $x=17; mutate($x); probe($x); }`, "?"},
		{`function consume($value){} function example(){ $x=17; consume($x); probe($x); }`, "17"},
		{`function example(){ $x=17; strlen($x); probe($x); }`, "17"},
		{`function example(){ $x=17; unknown($x); probe($x); }`, "?"},
		{`class Writer{function __construct(&$value){}} function example(){ $x=17; new Writer($x); probe($x); }`, "?"},
		{`class Reader{function __construct($value){}} function example(){ $x=17; new Reader($x); probe($x); }`, "17"},
		{`class Writer{function write(&$value){}} function example(Writer $w){ $x=17; $w->write($x); probe($x); }`, "?"},
		{`function mutate(&$value){} function example(){ $x=17; mutate(value:$x); probe($x); }`, "?"},
		{`function consume($value){} function example(){ $x=17; consume(value:$x); probe($x); }`, "17"},
		{`function probe($value){} function consume($value){} function example($again){$x=17; while($again){probe($x);consume($x);}}`, "17"},
		{`function mutate(&$value){} function example($again){$x=17; while($again){probe($x);mutate($x);}}`, "?"},
	} {
		t.Run(tc.code, func(t *testing.T) {
			if got := discoverOf(t, "<?php "+tc.code); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestUnlinkedArgumentCannotProveByValue(t *testing.T) {
	if !ArgumentMayWrite(nil, &syntax.Arg{}) {
		t.Fatal("unlinked argument was treated as by value")
	}
}
