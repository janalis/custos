package decryptionfailureusedasplaintext

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewSafety(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"DecryptionFailureUsedAsPlaintext"}})
	if err != nil {
		t.Fatal(err)
	}
	const decrypt = `<?php function f($c, $k, $path) { $p = openssl_decrypt($c, "aes-256-cbc", $k); `
	for _, tc := range []struct {
		name, body string
		findings   int
	}{
		{"assignment in condition", `if (($q = openssl_decrypt($c, "x", $k)) === false) { return; } strlen($q); }`, 0},
		{"and operand", `return $p !== false && strlen($p) > 3; }`, 0},
		{"ternary", `return $p === false ? -1 : strlen($p); }`, 0},
		{"negated truthiness exit", `if (!$p) { return; } file_put_contents($path, $p); }`, 0},
		{"is_string", `if (is_string($p)) { return strlen($p); } }`, 0},
		{"named is_string argument", `if (is_string(value: $p)) { return strlen($p); } }`, 0},
		{"earlier exit", `if (false === $p) { throw new RuntimeException(); } $n = 1; return strlen($p) + $n; }`, 0},
		// A predicate around the assignment and an exit through an endless loop
		// are proven by the local walk only.
		{"predicate assignment and operand", `return is_string($q = openssl_decrypt($c, "x", $k)) && strlen($q) > 3; }`, 0},
		{"predicate assignment ternary", `return is_string($q = openssl_decrypt($c, "x", $k)) ? strlen($q) : 0; }`, 0},
		{"predicate assignment elseif", `if ($path) { return 0; } elseif (is_string($q = openssl_decrypt($c, "x", $k))) { return strlen($q); } }`, 0},
		{"predicate assignment branch", `if (is_string($q = openssl_decrypt($c, "x", $k))) { return strlen($q); } }`, 0},
		{"predicate assignment else", `if (!is_string($q = openssl_decrypt($c, "x", $k))) { $n = 0; } else { $n = strlen($q); } }`, 0},
		{"predicate assignment after elseif", `if ($path === "") { $n = 1; } elseif (!is_string($q = openssl_decrypt($c, "x", $k))) { $n = 0; } else { $n = strlen($q); } }`, 0},
		{"predicate assignment exit", `if (!is_string($q = openssl_decrypt($c, "x", $k))) { return; } strlen($q); }`, 0},
		{"empty exit through endless loop", `if (empty($p)) { while (true) { sleep(1); } } return strlen($p); }`, 0},
		{"loose exit through endless loop", `if ($p == false) { while (true) { sleep(1); } } return strlen($p); }`, 0},
		{"predicate assignment failed branch", `if (!is_string($q = openssl_decrypt($c, "x", $k))) { return strlen($q); } }`, 1},
		{"other predicate", `if (is_scalar($p)) { return strlen($p); } }`, 1},
		{"reference assignment", `if (($p = &$k) === false) { return; } $p = openssl_decrypt($c, "x", $k); strlen($p); }`, 1},
		{"other local", `$q = $p; if ($q === false) { return; } strlen($p); }`, 1},
		{"variable variable", `if ($$k === false) { return; } strlen($p); }`, 1},
		{"non-terminating exit", `if ($p === false) { log(); } strlen($p); }`, 1},
		{"or operand reached on failure", `return $p !== false || strlen($p) > 3; }`, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.Parse("review.php", []byte(decrypt+tc.body), syntax.Options{})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
	// Parser recovery must neither panic nor invent a guard.
	broken := `<?php $p = openssl_decrypt($c,"x",$k); if ( { } elseif ( ) { strlen($p); } else strlen($p); $p ? : strlen($p); $p && ; strlen($p)`
	if got := e.Analyze(syntax.Parse("broken.php", []byte(broken), syntax.Options{})); len(got) == 0 {
		t.Fatal("broken source lost its unguarded finding")
	}
}
