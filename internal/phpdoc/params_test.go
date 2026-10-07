package phpdoc

import "testing"

func TestParamsNameFirst(t *testing.T) {
	d := Parse(`/**
 * @param int $a first
 * @param $b stdClass
 * @param $c string|null the c
 * @param $d the value
 * @param $e
 */`)
	want := []Param{{"int", "a"}, {"stdClass", "b"}, {"string|null", "c"}, {"", "d"}, {"", "e"}}
	got := d.Params()
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("param %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}
