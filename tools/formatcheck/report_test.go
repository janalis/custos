package main

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// The committed divergences load, and each names a rule of its preset's
// checker by its shape: a sniff code for PHP_CodeSniffer, a fixer name for
// PHP-CS-Fixer and Pint, or any rule within some files.
func TestCommittedDivergences(t *testing.T) {
	div, err := loadDivergences("../../internal/format/testdata/divergences.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range presetNames() {
		if _, ok := div[p]; !ok {
			t.Errorf("no entry for %s", p)
		}
	}
	for preset, ds := range div {
		sniffs := strings.HasPrefix(checkers[preset].label, "phpcs")
		for _, d := range ds {
			if d.Rule == "*" && d.File != "" {
				continue // every rule, in one kind of file
			}
			if strings.Contains(d.Rule, ".") != sniffs {
				t.Errorf("%s: %q does not look like a rule of %s", preset, d.Rule, checkers[preset].label)
			}
		}
	}
}

func TestLoadDivergencesErrors(t *testing.T) {
	for name, body := range map[string]string{
		"syntax":  "{",
		"preset":  `{"psr1": []}`,
		"pattern": `{"per": [{"rule": "[", "reason": "r"}]}`,
		"rule":    `{"per": [{"reason": "r"}]}`,
		"reason":  `{"per": [{"rule": "x"}]}`,
	} {
		file := filepath.Join(t.TempDir(), "d.json")
		write(t, file, body)
		if _, err := loadDivergences(file); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}

func TestSummarize(t *testing.T) {
	div := []divergence{{Rule: "phpdoc_*", Reason: "docblocks"}}
	before := []finding{{Rule: "b"}, {Rule: "a"}, {Rule: "a"}, {Rule: "phpdoc_trim"}}
	after := []finding{{Rule: "c", File: "y", Line: 2}, {Rule: "c", File: "x", Line: 9}, {Rule: "c", File: "x", Line: 1}, {Rule: "phpdoc_trim"}, {Rule: "a"}}
	oc := summarize("k", "psr12", checkers["psr12"], before, after, nil, div)
	if oc.Before != 4 || oc.After != 5 || oc.Divergent != 1 || oc.Remaining != 4 {
		t.Errorf("%+v", oc)
	}
	var rules []string
	for _, rc := range oc.Rules {
		rules = append(rules, rc.Rule+":"+rc.Divergent)
	}
	if want := []string{"c:", "a:", "phpdoc_trim:docblocks", "b:"}; !slices.Equal(rules, want) {
		t.Errorf("rules %q, want %q", rules, want)
	}
	if got := oc.Findings; got[0].Rule != "a" || got[1] != (finding{Rule: "c", File: "x", Line: 1}) || got[3].File != "y" {
		t.Errorf("findings %+v", got)
	}
	// With a gate, only what it found remains.
	gated := []finding{{Rule: "g"}}
	oc = summarize("k", "laravel", checkers["laravel"], before, after, gated, div)
	if oc.Remaining != 1 || oc.Findings[0].Rule != "g" || oc.Divergent != 1 {
		t.Errorf("%+v", oc)
	}
}

func TestPrintText(t *testing.T) {
	var b strings.Builder
	oc := outcome{Corpus: "k", Preset: "drupal", Checker: "phpcs", Remaining: 3, Findings: []finding{
		{Rule: "a", File: "f", Line: 1, Message: "m1"}, {Rule: "a", File: "f", Line: 2, Message: "m2"}, {Rule: "b", File: "g", Line: 3, Message: "m3"},
	}}
	printText(&b, []outcome{oc, {Corpus: "k", Preset: "psr12", Checker: "phpcs"}}, 1)
	want := "corpus  preset  checker  before  after  divergent  remaining\n" +
		"k       drupal  phpcs    0       0      0          3\n" +
		"k       psr12   phpcs    0       0      0          0\n" +
		"\nk drupal: 3 remaining\n  a: 2\n    f:1: m1\n  b: 1\n    g:3: m3\n"
	if b.String() != want {
		t.Errorf("got\n%s\nwant\n%s", b.String(), want)
	}
}

// A divergence limited to a file covers only the findings of that file.
func TestFileDivergence(t *testing.T) {
	div := []divergence{{Rule: "statement_indentation", File: "Process/Fake.php", Reason: "quirk"}}
	after := []finding{
		{Rule: "statement_indentation", File: "src/Process/Fake.php", Line: 3},
		{Rule: "statement_indentation", File: "src/Other.php", Line: 4},
	}
	oc := summarize("k", "psr12", checkers["psr12"], nil, after, nil, div)
	if oc.Divergent != 1 || oc.Remaining != 1 || oc.Findings[0].File != "src/Other.php" || oc.Rules[0].Divergent != "" {
		t.Fatalf("plain checker: %+v", oc)
	}
	oc = summarize("k", "per", checkers["per"], nil, after, after, div)
	if oc.Remaining != 1 || oc.Findings[0].File != "src/Other.php" {
		t.Fatalf("gated checker: %+v", oc)
	}
	if !inFile(div, finding{Rule: "a,statement_indentation", File: "src/Process/Fake.php"}) {
		t.Fatal("a fixer list naming the rule")
	}
	if got := ruleWide(append(div, divergence{Rule: "x", Reason: "y"})); len(got) != 1 || got[0].Rule != "x" {
		t.Fatalf("rule-wide: %+v", got)
	}
}
