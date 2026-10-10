package hardcodedcredentialatknownsink

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestCredentialPathPolicy(t *testing.T) {
	for _, tc := range []struct {
		path     string
		patterns []string
		want     int
	}{
		{"/project/tests/example.php", []string{"**/tests/**"}, 0},
		{"/project/app.php", []string{"**/tests/**"}, 1},
		{"/project/example.php", []string{"[invalid"}, 1},
		{"/project/example.php", nil, 1},
		{strings.Repeat("a/", 257) + "example.php", []string{"**/tests/**"}, 1},
		{"app.php", []string{strings.Repeat("**/", 129) + "absent"}, 1},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"HardcodedCredentialAtKnownSink"}, Rules: map[string]analysis.RuleConfig{"HardcodedCredentialAtKnownSink": {Options: map[string]any{"excludedPaths": tc.patterns}}}})
		if err != nil {
			t.Fatal(err)
		}
		f := syntax.Parse(tc.path, []byte(`<?php new PDO($dsn,"service","password");`), syntax.Options{})
		if got := e.Analyze(f); len(got) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.path, len(got), tc.want)
		}
	}
}

func BenchmarkCredentialPathPolicy(b *testing.B) {
	pattern := strings.Split(strings.Repeat("**/", 120)+"tests/*.php", "/")
	parts := strings.Split(strings.Repeat("a/", 200)+"app.php", "/")
	b.ReportAllocs()
	for b.Loop() {
		if glob(pattern, parts) {
			b.Fatal("unexpected match")
		}
	}
}
