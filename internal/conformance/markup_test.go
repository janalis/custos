package conformance

import (
	"testing"

	"custos/internal/meta"
)

func TestParseMarkup(t *testing.T) {
	src := []byte(`<?php
$a = <warning descr="outer &quot;x&quot;">foo(<weak_warning descr="inner">$b</weak_warning>)</warning>;
if ($x <error>< 1</error>) {}<caret>
echo "<b>";`)
	clean, exp, err := ParseMarkup(src)
	if err != nil {
		t.Fatal(err)
	}
	wantClean := "<?php\n$a = foo($b);\nif ($x < 1) {}\necho \"<b>\";"
	if string(clean) != wantClean {
		t.Fatalf("clean mismatch:\n%q\n%q", clean, wantClean)
	}
	want := []Expectation{
		{meta.SeverityWarning, `outer "x"`, 11, 18},
		{meta.SeverityInfo, "inner", 15, 17},
		{meta.SeverityError, "", 27, 30},
	}
	if len(exp) != len(want) {
		t.Fatalf("got %v", exp)
	}
	for i := range want {
		if exp[i] != want[i] {
			t.Errorf("#%d: got %s want %s", i, exp[i], want[i])
		}
	}
}

func TestParseMarkupUnbalanced(t *testing.T) {
	for _, src := range []string{"<warning>x", "x</warning>", "<warning>x</error>"} {
		if _, _, err := ParseMarkup([]byte(src)); err == nil {
			t.Errorf("%q: expected error", src)
		}
	}
}

func TestDiffCountsDuplicates(t *testing.T) {
	a := Expectation{meta.SeverityWarning, "m", 1, 2}
	missing, unexpected := diff([]Expectation{a, a}, []Expectation{a}, true)
	if len(missing) != 1 || len(unexpected) != 0 {
		t.Fatalf("missing=%v unexpected=%v", missing, unexpected)
	}
}
