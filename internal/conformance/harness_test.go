package conformance

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"custos/internal/meta"
)

// fakeEngine returns canned results, to exercise Check without rules.
type fakeEngine struct {
	findings []Finding
	fixed    string
	err      error
	fixErr   error
}

func (f *fakeEngine) Supports([]string) bool { return true }
func (f *fakeEngine) Analyze(Request) ([]Finding, error) {
	return f.findings, f.err
}
func (f *fakeEngine) Fix(Request) ([]byte, error) { return []byte(f.fixed), f.fixErr }

func TestCheck(t *testing.T) {
	marked := []byte(`<?php <warning descr="[EA] m">$a</warning>; <error descr="Expected: ;">$b</error>`)
	eng := &fakeEngine{findings: []Finding{{Rule: "R", Severity: meta.SeverityWarning, Message: "other", Start: 6, End: 8}}, fixed: "<?php  $x;  $b"}

	// EA mode: parse markers without the prefix are dropped, messages are
	// not compared, the expected fix is hidden.
	res, err := Check(eng, Request{}, marked, []byte("<?php $a; $b"), CompareOptions{OnlyPrefixed: "[EA] ", HideExpected: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Missing) != 0 || len(res.Unexpected) != 0 || res.OK() {
		t.Fatalf("EA mode: %+v", res)
	}
	if !strings.Contains(res.FixDiff, "collapsed offset 7") || strings.Contains(res.FixDiff, "$a") {
		t.Fatalf("hidden fix diff: %q", res.FixDiff)
	}

	// own fixtures: messages compared, full diff shown
	res, err = Check(eng, Request{}, marked, []byte("<?php $a; $b"), CompareOptions{Messages: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Missing) != 2 || len(res.Unexpected) != 1 || !strings.Contains(res.FixDiff, "--- want\n<?php $a; $b\n--- got") {
		t.Fatalf("own mode: %+v", res)
	}
	if res.Missing[0].String() != `warning[6:8] "[EA] m"` {
		t.Fatalf("expectation string: %s", res.Missing[0])
	}

	// whitespace-only fix differences pass
	eng.fixed = "<?php\n$a;\n\t$b\n"
	if res, err := Check(eng, Request{}, []byte("<?php $a; $b"), []byte("<?php $a; $b"), CompareOptions{}); err != nil || res.FixDiff != "" {
		t.Fatalf("whitespace fix: %+v %v", res, err)
	}

	if _, err := Check(eng, Request{}, []byte("<warning>"), nil, CompareOptions{}); err == nil {
		t.Fatal("bad markup accepted")
	}
	boom := errors.New("boom")
	if _, err := Check(&fakeEngine{err: boom}, Request{}, []byte("x"), nil, CompareOptions{}); !errors.Is(err, boom) {
		t.Fatalf("analyze error: %v", err)
	}
	if _, err := Check(&fakeEngine{fixErr: boom}, Request{}, []byte("x"), []byte("x"), CompareOptions{}); !errors.Is(err, boom) {
		t.Fatalf("fix error: %v", err)
	}
}

func TestSortExpectations(t *testing.T) {
	e := []Expectation{
		{Severity: meta.SeverityWarning, Start: 1, End: 3, Message: "b"},
		{Severity: meta.SeverityWarning, Start: 1, End: 3, Message: "a"},
		{Severity: meta.SeverityError, Start: 1, End: 3, Message: "z"},
		{Severity: meta.SeverityError, Start: 1, End: 2},
		{Severity: meta.SeverityError, Start: 0, End: 9},
	}
	SortExpectations(e)
	var got []string
	for _, x := range e {
		got = append(got, x.String())
	}
	want := `error[0:9] "",error[1:2] "",error[1:3] "z",warning[1:3] "a",warning[1:3] "b"`
	if strings.Join(got, ",") != want {
		t.Fatalf("got %s", strings.Join(got, ","))
	}
}

// TestMarkupNotTags: text that only looks like a tag stays in the source.
func TestMarkupNotTags(t *testing.T) {
	for _, src := range []string{
		`<warningx>`,            // longer name
		`<warning descr="open>`, // unterminated attribute
		`<warning descr="x"`,    // no closing '>'
		`<warning`,              // truncated
	} {
		clean, exp, err := ParseMarkup([]byte(src))
		if err != nil || string(clean) != src || len(exp) != 0 {
			t.Errorf("%q: clean=%q exp=%v err=%v", src, clean, exp, err)
		}
	}
	clean, exp, err := ParseMarkup([]byte(`<info textAttributesKey="k">x</info>`))
	if err != nil || string(clean) != "x" || len(exp) != 1 || exp[0].Message != "" || exp[0].Severity != meta.SeverityInfo {
		t.Fatalf("attribute without descr: %q %v %v", clean, exp, err)
	}
}

func TestIndexFiles(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "index.json")
	if _, err := LoadEAIndex(p); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing index: %v", err)
	}
	if err := os.WriteFile(p, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadEAIndex(p); err == nil || !strings.Contains(err.Error(), "index.json") {
		t.Fatalf("invalid index: %v", err)
	}
	if err := os.WriteFile(p, []byte(`{"eaPath": "`+filepath.ToSlash(dir)+`", "cases": [{"test": "T", "rules": ["R"], "fixture": "f.php"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	idx, err := LoadEAIndex(p)
	if err != nil || len(idx.Cases) != 1 || idx.Cases[0].Fixture != "f.php" {
		t.Fatalf("index: %+v %v", idx, err)
	}
	if err := os.WriteFile(filepath.Join(dir, "f.php"), []byte("<?php"), 0o644); err != nil {
		t.Fatal(err)
	}
	if b, err := ReadEA(idx, "f.php"); err != nil || string(b) != "<?php" {
		t.Fatalf("ReadEA: %q %v", b, err)
	}
}

func TestExcerpt(t *testing.T) {
	b := []byte(strings.Repeat("a", 50) + "X" + strings.Repeat("b", 100))
	if got := excerpt(b, 50); got != strings.Repeat("a", 40)+"X"+strings.Repeat("b", 59) {
		t.Fatalf("excerpt: %q", got)
	}
	if got := excerpt([]byte("ab"), 1); got != "ab" {
		t.Fatalf("short excerpt: %q", got)
	}
}
