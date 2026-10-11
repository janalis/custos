package hostcookieprefixcontractviolation

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, source string
		findings     int
	}{
		{"capitalised keys", `<?php setcookie("__Host-a", "t", ["Secure" => true, "Path" => "/", "HttpOnly" => true]);`, 0},
		{"upper keys", `<?php setcookie("__Host-a", "t", ["SECURE" => true, "PATH" => "/"]);`, 0},
		{"capitalised bad path", `<?php setcookie("__Host-a", "t", ["Secure" => true, "Path" => "/app"]);`, 1},
		{"capitalised insecure", `<?php setcookie("__Host-a", "t", ["Secure" => false, "Path" => "/"]);`, 1},
		{"capitalised domain", `<?php setcookie("__Host-a", "t", ["secure" => true, "path" => "/", "DOMAIN" => "example.test"]);`, 1},
		{"two spellings", `<?php setcookie("__Host-a", "t", ["secure" => true, "path" => "/app", "Path" => "/"]);`, 0},
		{"two spellings reversed", `<?php setcookie("__Host-a", "t", ["secure" => true, "Path" => "/", "path" => "/app"]);`, 0},
		{"same length other key", `<?php setcookie("__Host-a", "t", ["secure" => true, "path" => "/", "damain" => "example.test"]);`, 0},
		{"non-ascii lookalike", "<?php setcookie(\"__Host-a\", \"t\", [\"ſecure\" => true, \"path\" => \"/\"]);", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			opt := syntax.Options{Version: phpversion.PHP85}
			f := syntax.ParseBest("review.php", []byte(tc.source), opt)
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
