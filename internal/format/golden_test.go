package format

import (
	"flag"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/janalis/custos/internal/php/syntax"
)

var update = flag.Bool("update", false, "rewrite the expected outputs of the golden tests")

// TestGolden formats every testdata/cases/<name>.php with the settings its
// first line comment names and compares the output byte for byte with
// <name>.<preset>.php. The input's first line is
//
//	<?php // presets: per symfony; set: line_length=60, wrap=long; risky
func TestGolden(t *testing.T) {
	inputs, _ := filepath.Glob("testdata/cases/*.php")
	n := 0
	for _, in := range inputs {
		if strings.Count(filepath.Base(in), ".") > 1 {
			continue // an expected output
		}
		n++
		src, err := os.ReadFile(in)
		if err != nil {
			t.Fatal(err)
		}
		presets, layer, risky := directives(t, string(src))
		for _, p := range presets {
			r, err := Resolve(p, risky, layer)
			if err != nil {
				t.Fatalf("%s: %v", in, err)
			}
			res := Source(in, src, Options{Settings: &r.Settings})
			want := strings.TrimSuffix(in, ".php") + "." + p + ".php"
			got := res.Output
			if res.Skipped != "" {
				got = []byte("skipped: " + res.Skipped + "\n")
			}
			if *update {
				if err := os.WriteFile(want, got, 0o644); err != nil {
					t.Fatal(err)
				}
				continue
			}
			exp, err := os.ReadFile(want)
			if err != nil {
				t.Fatalf("%s: %v (run go test -update)", in, err)
			}
			if string(exp) != string(got) {
				t.Errorf("%s with %s:\n--- want\n%s\n--- got\n%s", in, p, exp, got)
			}
			if res.Skipped == "" {
				checkResult(t, in+" "+p, src, res, &r.Settings)
				checkGroups(t, in+" "+p, src, res)
			}
		}
	}
	if n == 0 {
		t.Fatal("no golden case")
	}
}

// checkResult asserts the properties every formatting result has: the
// changes replay to the output, and formatting the output changes nothing.
func checkResult(t *testing.T, name string, src []byte, res Result, s *Settings) {
	t.Helper()
	if got := apply(src, res.Changes); string(got) != string(res.Output) {
		t.Errorf("%s: changes replay to\n%s", name, got)
	}
	again := Source("again.php", res.Output, Options{Settings: s})
	for _, c := range again.Changes {
		if !c.Report {
			t.Errorf("%s: not idempotent: %s %q at %d", name, c.Rule, c.New, c.Start)
			break
		}
	}
}

// directives reads the test settings from the first line of a case.
func directives(t *testing.T, src string) ([]string, Layer, bool) {
	t.Helper()
	var spec string
	ok := false
	for _, line := range strings.SplitN(strings.ReplaceAll(src, "\r", "\n"), "\n", 3)[:2] {
		if _, spec, ok = strings.Cut(line, "// presets"); ok {
			spec = "presets" + spec
			break
		}
	}
	if !ok {
		t.Fatalf("no directives in %.40q", src)
	}
	var presets []string
	layer := Layer{Source: "test", Values: map[string]any{}}
	risky := false
	for _, part := range strings.Split(spec, ";") {
		part = strings.TrimSpace(part)
		key, val, _ := strings.Cut(part, ":")
		switch strings.TrimSpace(key) {
		case "presets":
			presets = strings.Fields(val)
		case "set":
			for _, kv := range strings.Split(val, ",") {
				k, v, _ := strings.Cut(strings.TrimSpace(kv), "=")
				d, ok := Lookup(k)
				if !ok {
					t.Fatalf("unknown setting %q", k)
				}
				if d.Kind == KindList {
					layer.Values[k] = strings.Split(strings.ReplaceAll(v, "|", ","), ",")
					if v == "" {
						layer.Values[k] = []string{}
					}
				} else {
					layer.Values[k] = v
				}
			}
		case "risky":
			risky = true
		}
	}
	if len(presets) == 0 {
		presets = Presets()
	}
	sort.Strings(presets)
	return presets, layer, risky
}

// checkGroups asserts that every change applied alone, or every group of
// changes applied together, leaves code that still parses: what an editor
// gets when it takes one quick-fix or formats a range.
func checkGroups(t *testing.T, name string, src []byte, res Result) {
	t.Helper()
	before := len(syntax.ParseBest(name, src, syntax.Options{}).Errors)
	groups := map[int][]Change{}
	for i, c := range res.Changes {
		key := c.Group
		if key == 0 {
			key = -1 - i
		}
		if !c.Report {
			groups[key] = append(groups[key], c)
		}
	}
	for key, cs := range groups {
		out := apply(src, cs)
		if n := len(syntax.ParseBest(name, out, syntax.Options{}).Errors); n > before {
			t.Errorf("%s: applying group %d alone (%s) breaks the syntax:\n%s", name, key, cs[0].Rule, out)
		}
	}
}
