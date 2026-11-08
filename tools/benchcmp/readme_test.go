//go:build unix

package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestReadmeSection(t *testing.T) {
	r := sampleResults()
	r.Corpora = append([]CorpusResult{{Name: "small", Repo: "https://github.com/v/small.git", Tag: "1", Files: 3}}, r.Corpora...)
	r.Corpora[1].Files = 12345
	got := readmeSection(r, sampleFeatures())
	for _, want := range []string{
		"on [v/app](https://github.com/v/app) `v1.0.0` (12,345 PHP files), 4 cores, median of 3 runs:",
		"| **custos analyse** | analyzer | 0.50 s | 80 MB |\n| Mid |",
		"| Broken | analyzer | error: &lt;bad&gt; \\| &#123;&#123;x&#125;&#125; more | — |",
		"| Quick | linter | 0.02 s | 20 MB |",
		"| Slow | analyzer | 150 s | 2.0 GB |",
		"✓ yes, ◐ partial, — no",
		"| Fixes | ✓ | ◐ |",
		"(" + pageURL + ")",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if got := readmeSection(&Results{}, nil); strings.Contains(got, "Speed") || strings.Contains(got, "Features") || !strings.Contains(got, pageURL) {
		t.Errorf("no results, no features:\n%s", got)
	}
}

func TestThousands(t *testing.T) {
	for n, want := range map[int]string{0: "0", 999: "999", 1000: "1,000", 1746: "1,746", 1234567: "1,234,567"} {
		if got := thousands(n); got != want {
			t.Errorf("thousands(%d) = %q, want %q", n, got, want)
		}
	}
}

// The committed README holds the comparison bench/results.json and
// bench/features.json render.
func TestReadmeComparisonIsCurrent(t *testing.T) {
	f, err := loadFeatures("../../bench/features.json")
	if err != nil {
		t.Fatal(err)
	}
	var r Results
	if err := json.Unmarshal([]byte(read(t, "../../bench/results.json")), &r); err != nil {
		t.Fatal(err)
	}
	readme := read(t, "../../README.md")
	if want, err := spliceReadme(readme, readmeSection(&r, f)); err != nil || want != readme {
		t.Errorf("README comparison is stale (make compare-doc): %v", err)
	}
}
