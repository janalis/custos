package util

import "testing"

func TestStringLiteralValue(t *testing.T) {
	for raw, want := range map[string]string{
		`'abc'`:          "abc",
		`'\\Lib\\fmt'`:   `\Lib\fmt`,
		`'it\'s \n'`:     `it's \n`,
		`"a\tb"`:         "a\tb",
		`"\\Lib\\fmt"`:   `\Lib\fmt`,
		`"\x41\101\$\q"`: `AA$\q`,
		`"\u{48}i"`:      "Hi",
	} {
		got, ok := StringLiteralValue(raw)
		if !ok || got != want {
			t.Errorf("%s: got %q, %v; want %q", raw, got, ok, want)
		}
	}
	if _, ok := StringLiteralValue("12"); ok {
		t.Error("number accepted")
	}
}

func TestStringLiteralValueEdges(t *testing.T) {
	if _, ok := StringLiteralValue("1231"); ok {
		t.Fatal("number with equal first/last byte")
	}
	if v, ok := StringLiteralValue(`"\r\v\e\f"`); !ok || v != "\r\v\x1b\f" {
		t.Fatalf("got %q %v", v, ok)
	}
}
