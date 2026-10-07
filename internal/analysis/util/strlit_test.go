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
