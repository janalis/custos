package baseline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"custos/internal/diagnostic"
	"custos/internal/php/syntax"
)

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	php := filepath.Join(dir, "a.php")
	if err := os.WriteFile(php, []byte("<?php\n$a = !!$b;\n$c = !!$d;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	items := []diagnostic.Item{
		{Path: php, Line: 2, Rule: "NestedNotOperators", Message: "m"},
		{Path: php, Line: 3, Rule: "NestedNotOperators", Message: "m"},
	}
	bl := filepath.Join(dir, "custos-baseline.json")
	if n, err := Write(bl, items); err != nil || n != 2 {
		t.Fatalf("write: %d %v", n, err)
	}
	// Code moves down two lines and a new finding appears.
	if err := os.WriteFile(php, []byte("<?php\n// x\n// y\n$a = !!$b;\n$c = !!$d;\n$e = !!$f;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	moved := []diagnostic.Item{
		{Path: php, Line: 4, Rule: "NestedNotOperators", Message: "m"},
		{Path: php, Line: 5, Rule: "NestedNotOperators", Message: "m"},
		{Path: php, Line: 6, Rule: "NestedNotOperators", Message: "m"},
	}
	b, err := Load(bl)
	if err != nil {
		t.Fatal(err)
	}
	rest, suppressed := b.Filter(moved)
	if suppressed != 2 || len(rest) != 1 || rest[0].Line != 6 {
		t.Fatalf("suppressed=%d rest=%+v", suppressed, rest)
	}
}

func TestLoadErrors(t *testing.T) {
	dir := t.TempDir()
	if _, err := Load(filepath.Join(dir, "missing.json")); err == nil {
		t.Fatal("missing baseline accepted")
	}
	bad := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(bad, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(bad); err == nil || !strings.Contains(err.Error(), "bad.json") {
		t.Fatalf("invalid JSON: %v", err)
	}
}

// TestWriteSortsAndSkips: tool findings (syntax/internal/io) are never
// baselined; entries are sorted by path, rule then line text; a line past
// the end of the file (or an unreadable file) keys on empty text.
func TestWriteSortsAndSkips(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.php")
	b := filepath.Join(dir, "b.php")
	for _, p := range []string{a, b} {
		if err := os.WriteFile(p, []byte("<?php\n$x = 1;\n$y = 2;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	items := []diagnostic.Item{
		{Path: b, Line: 2, Rule: "R1", Message: "m"},
		{Path: a, Line: 3, Rule: "R2", Message: "m"},
		{Path: a, Line: 3, Rule: "R1", Message: "m"},
		{Path: a, Line: 2, Rule: "R1", Message: "m"},
		{Path: a, Line: 99, Rule: "R1", Message: "m"},
		{Path: a, Line: 1, Rule: "syntax", Message: "m"},
		{Path: a, Line: 1, Rule: "internal", Message: "m"},
		{Path: a, Line: 1, Rule: "io", Message: "m"},
	}
	bl := filepath.Join(dir, "bl.json")
	if n, err := Write(bl, items); err != nil || n != 5 {
		t.Fatalf("write: %d %v", n, err)
	}
	data, err := os.ReadFile(bl)
	if err != nil {
		t.Fatal(err)
	}
	var f File
	if err := json.Unmarshal(data, &f); err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, e := range f.Entries {
		got = append(got, e.Path+"|"+e.Rule+"|"+e.Line)
	}
	want := []string{"a.php|R1|", "a.php|R1|$x = 1;", "a.php|R1|$y = 2;", "a.php|R2|$y = 2;", "b.php|R1|$x = 1;"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("entries:\n got %v\nwant %v", got, want)
	}
	if _, err := Write(filepath.Join(dir, "no", "such", "dir.json"), items); err == nil {
		t.Fatal("write into a missing directory succeeded")
	}
}

// TestEmptyFiltersNothing: the empty baseline has no directory, so paths
// stay absolute and nothing is suppressed.
func TestEmptyFiltersNothing(t *testing.T) {
	items := []diagnostic.Item{{Path: filepath.Join(t.TempDir(), "a.php"), Line: 1, Rule: "R", Message: "m"}}
	rest, n := Empty().Filter(items)
	if n != 0 || len(rest) != 1 {
		t.Fatalf("suppressed=%d rest=%v", n, rest)
	}
	if got := rel("", "/abs/x.php"); got != "/abs/x.php" {
		t.Fatalf("rel without dir: %q", got)
	}
}

// TestNoWorkingDirectory: when the working directory cannot be resolved
// (unreadable and removed, $PWD unset) relative paths cannot be made
// absolute: Write reports it, and Filter keys on the path as given. getcwd
// fails on macOS for an unreadable directory and on Linux (even as root)
// for a removed one, so the test does both.
func TestNoWorkingDirectory(t *testing.T) {
	sub := filepath.Join(t.TempDir(), "cwd")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(sub)
	t.Setenv("PWD", "")
	if err := os.Chmod(sub, 0); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(sub); err != nil {
		t.Fatal(err)
	}
	if _, err := Write("bl.json", nil); err == nil {
		t.Fatal("write with no working directory succeeded")
	}
	if got := rel("/x", "a.php"); got != "a.php" {
		t.Fatalf("rel: %q", got)
	}
}

// TestFilterFindings: one file's findings are matched on the buffer's line
// text, numbered and split as Filter sees them (a lone CR ends a line for
// positions but not for line text, so such a line keys as "").
func TestFilterFindings(t *testing.T) {
	dir := t.TempDir()
	bl := filepath.Join(dir, "bl.json")
	if err := os.WriteFile(bl, []byte(`{"version": 1, "entries": [
		{"path": "a.php", "rule": "R", "message": "m", "line": "$a;", "count": 1},
		{"path": "a.php", "rule": "R", "message": "m", "line": "", "count": 1}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	b, err := Load(bl)
	if err != nil {
		t.Fatal(err)
	}
	php := filepath.Join(dir, "a.php")
	src := []byte("<?php\n$a;\n $a;\r$b;")
	at := func(off uint32) diagnostic.Finding {
		return diagnostic.Finding{Rule: "R", Message: "m", Span: syntax.Span{Start: off, End: off + 1}}
	}
	rest := b.FilterFindings(php, src, []diagnostic.Finding{at(6), at(11), at(15), at(15)})
	if len(rest) != 2 || rest[0].Span.Start != 11 || rest[1].Span.Start != 15 {
		t.Fatalf("rest: %+v", rest)
	}
	if got := Empty().FilterFindings(php, src, []diagnostic.Finding{at(6)}); len(got) != 1 {
		t.Fatalf("empty baseline: %+v", got)
	}
	// A file without entries keeps its findings.
	if got := b.FilterFindings(filepath.Join(dir, "b.php"), src, []diagnostic.Finding{at(6)}); len(got) != 1 {
		t.Fatalf("file without entries: %+v", got)
	}
}

// BenchmarkFilterFindings measures the per-analysis cost the language server
// pays for a file the baseline records and for one it does not.
func BenchmarkFilterFindings(b *testing.B) {
	dir := b.TempDir()
	path := filepath.Join(dir, "bl.json")
	if err := os.WriteFile(path, []byte(`{"version": 1, "entries": [{"path": "a.php", "rule": "R", "message": "m", "line": "$a;", "count": 1}]}`), 0o644); err != nil {
		b.Fatal(err)
	}
	bl, err := Load(path)
	if err != nil {
		b.Fatal(err)
	}
	src := []byte("<?php\n" + strings.Repeat("$a;\n", 5000))
	findings := []diagnostic.Finding{{Rule: "R", Message: "m", Span: syntax.Span{Start: 6, End: 9}}}
	for _, name := range []string{"a.php", "b.php"} {
		file := filepath.Join(dir, name)
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				bl.FilterFindings(file, src, findings)
			}
		})
	}
}

// TestMovedFindings: a finding that moved between rules related through
// the catalogue's suppressedBy (NullPointerException leaves possible nulls
// to PossiblyNullAccess and PossiblyNullArgument) stays covered by the
// entry of the rule it moved from at the same location, whatever the
// message, in both directions. Exact entries are consumed first, unrelated
// rules and other lines never match, and each entry covers one finding.
func TestMovedFindings(t *testing.T) {
	dir := t.TempDir()
	php := filepath.Join(dir, "a.php")
	if err := os.WriteFile(php, []byte("<?php\n$a->b();\nf($c);\n$d->e();\n$g->h();\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	bl := filepath.Join(dir, "bl.json")
	if err := os.WriteFile(bl, []byte(`{"version": 1, "entries": [
		{"path": "a.php", "rule": "NullPointerException", "message": "old", "line": "$a->b();", "count": 1},
		{"path": "a.php", "rule": "NullPointerException", "message": "old", "line": "f($c);", "count": 1},
		{"path": "a.php", "rule": "PossiblyNullAccess", "message": "new", "line": "$d->e();", "count": 1},
		{"path": "a.php", "rule": "NullPointerException", "message": "old", "line": "$g->h();", "count": 1}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	b, err := Load(bl)
	if err != nil {
		t.Fatal(err)
	}
	items := []diagnostic.Item{
		{Path: php, Line: 2, Rule: "PossiblyNullAccess", Message: "new"},   // moved from NullPointerException
		{Path: php, Line: 2, Rule: "PossiblyNullAccess", Message: "new"},   // second one: the entry is used up
		{Path: php, Line: 3, Rule: "PossiblyNullArgument", Message: "new"}, // moved from NullPointerException
		{Path: php, Line: 4, Rule: "NullPointerException", Message: "old"}, // moved back
		{Path: php, Line: 3, Rule: "NestedNotOperators", Message: "old"},   // unrelated rule
		{Path: php, Line: 5, Rule: "PossiblyNullAccess", Message: "new"},   // the exact finding comes after
		{Path: php, Line: 5, Rule: "NullPointerException", Message: "old"}, // exact match wins the entry
	}
	rest, n := b.Filter(items)
	if n != 4 || len(rest) != 3 || rest[0].Line != 2 || rest[1].Rule != "NestedNotOperators" || rest[2].Line != 5 || rest[2].Rule != "PossiblyNullAccess" {
		t.Fatalf("suppressed=%d rest=%+v", n, rest)
	}

	src, _ := os.ReadFile(php)
	at := func(rule string, off uint32) diagnostic.Finding {
		return diagnostic.Finding{Rule: rule, Message: "new", Span: syntax.Span{Start: off, End: off + 1}}
	}
	fs := []diagnostic.Finding{
		at("PossiblyNullAccess", 6), at("PossiblyNullAccess", 6), at("PossiblyNullArgument", 15),
		at("NullPointerException", 22), at("PossiblyNullAccess", 31),
		{Rule: "NullPointerException", Message: "old", Span: syntax.Span{Start: 31, End: 32}},
	}
	got := b.FilterFindings(php, src, fs)
	if len(got) != 2 || got[0].Span.Start != 6 || got[1].Span.Start != 31 || got[1].Rule != "PossiblyNullAccess" {
		t.Fatalf("FilterFindings: %+v", got)
	}
}

// TestRecorder: a file's recorder tells which findings the baseline
// records without consuming entries: of the finding's rule and message, or
// of a related rule whatever the message, at its line text taken as
// FilterFindings takes it (CRLF, LF and lone CR line endings alike). A
// relative directory matches relative paths as they are; a file without
// entries has no recorder.
func TestRecorder(t *testing.T) {
	b, err := Parse([]byte(`{"version": 1, "entries": [
		{"path": "a.php", "rule": "R", "message": "m", "line": "$a;", "count": 1},
		{"path": "a.php", "rule": "R", "message": "m", "line": "", "count": 1},
		{"path": "a.php", "rule": "NullPointerException", "message": "other", "line": "$n->x();", "count": 1},
		{"path": "cr.php", "rule": "R", "message": "m", "line": "<?php", "count": 1}]}`), "src")
	if err != nil {
		t.Fatal(err)
	}
	at := func(rule, msg string, off uint32) diagnostic.Finding {
		return diagnostic.Finding{Rule: rule, Message: msg, Span: syntax.Span{Start: off, End: off + 1}}
	}
	type check struct {
		f    diagnostic.Finding
		want bool
	}
	for _, c := range []struct {
		path   string
		src    string
		checks []check
	}{
		{"src/a.php", "<?php\r\n$a;\n$n->x();\n $a;\r$b;", []check{
			{at("R", "m", 7), true},
			{at("R", "m", 8), true}, // not consumed by the first
			{at("R", "other", 7), false},
			{at("PossiblyNullAccess", "m", 11), true},
			{at("UnnecessarySemicolon", "m", 11), false},
			{at("R", "m", 20), false}, // the line runs to the newline: " $a;\r$b;"
			{at("R", "m", 25), true},  // after a lone CR: no such line, keyed ""
		}},
		// CR-only line endings: the first line holds the whole file, the
		// others key as "".
		{"src/cr.php", "<?php\r$a;\r$b;", []check{
			{at("R", "m", 0), false}, // the line is "<?php\r$a;\r$b;"
			{at("R", "m", 6), false}, // no entry of cr.php keys ""
		}},
	} {
		rec := b.Recorder(c.path, []byte(c.src))
		for _, k := range c.checks {
			if got := rec(k.f, nil); got != k.want {
				t.Errorf("%s: recorder(%+v) = %v, want %v", c.path, k.f, got, k.want)
			}
			// The same key as FilterFindings, with an entry left.
			if filtered := len(b.FilterFindings(c.path, []byte(c.src), []diagnostic.Finding{k.f})) == 0; filtered != k.want {
				t.Errorf("%s: FilterFindings(%+v) covered = %v, recorder %v", c.path, k.f, filtered, k.want)
			}
		}
	}
	// Earlier findings at the same line text take the entry's count first,
	// as FilterFindings consumes it; one at another line text does not; an
	// entry of a related rule covers whatever its count.
	src := []byte("<?php\n$a;\n$n->x();\n$a;\n")
	rec := b.Recorder("src/a.php", src)
	if rec(at("R", "m", 19), []diagnostic.Finding{at("R", "m", 6)}) {
		t.Error("count taken by an earlier finding at the same line text")
	}
	if !rec(at("R", "m", 19), []diagnostic.Finding{at("R", "m", 10)}) {
		t.Error("count taken by an earlier finding at another line text")
	}
	if !rec(at("PossiblyNullAccess", "m", 10), []diagnostic.Finding{at("PossiblyNullAccess", "m", 10)}) {
		t.Error("related entry with its count taken")
	}
	if covered := b.Covered("src/a.php", src, []diagnostic.Finding{at("R", "m", 6), at("R", "m", 19)}); !covered[0] || covered[1] {
		t.Errorf("FilterFindings consumes in source order: %v", covered)
	}
	if b.Recorder("src/b.php", []byte("<?php")) != nil {
		t.Fatal("recorder of a file without entries")
	}
	if _, err := Parse([]byte("{"), "."); err == nil {
		t.Fatal("invalid baseline parsed")
	}
}

// BenchmarkRecorder measures the questions a check asks about many
// findings of one large file the baseline records: the file's lines are
// indexed once, not once per finding.
func BenchmarkRecorder(b *testing.B) {
	bl, err := Parse([]byte(`{"version": 1, "entries": [{"path": "a.php", "rule": "R", "message": "m", "line": "$a;", "count": 1}]}`), ".")
	if err != nil {
		b.Fatal(err)
	}
	const lines = 50000
	src := []byte("<?php\n" + strings.Repeat("$a;\n", lines))
	b.ReportAllocs()
	for b.Loop() {
		rec := bl.Recorder("a.php", src)
		for i := range uint32(lines) {
			rec(diagnostic.Finding{Rule: "R", Message: "m", Span: syntax.Span{Start: 6 + 4*i, End: 9 + 4*i}}, nil)
		}
	}
}
