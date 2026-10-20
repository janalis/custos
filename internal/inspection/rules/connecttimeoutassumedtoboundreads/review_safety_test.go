package connecttimeoutassumedtoboundreads

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
	const open = `<?php function f($h) { $s = fsockopen($h, 80, $errno, $errstr, 2); `
	for _, tc := range []struct {
		name, source string
		findings     int
	}{
		{"early exit", open + `if ($s === false || !stream_set_timeout($s, 1)) { return null; } fwrite($s, "x"); return fgets($s); }`, 0},
		{"early exit comparison", open + `if (!$s || stream_set_timeout($s, 1) === false) { throw new \Exception(); } return fgets($s); }`, 0},
		{"early exit yoda", open + `if (true !== stream_set_timeout($s, 1)) { return null; } return fgets($s); }`, 0},
		{"early exit loose", open + `if (stream_set_timeout($s, 1) != true) { return null; } return fgets($s); }`, 0},
		{"early exit with else", open + `if (!stream_set_timeout($s, 1)) { return null; } else { fwrite($s, "x"); } return fgets($s); }`, 0},
		{"early exit in loop", open + `while (true) { if (!stream_set_timeout($s, 1)) { break; } fgets($s); } }`, 0},
		{"early exit outer list", open + `if (!stream_set_timeout($s, 1)) { return null; } if ($h) { return fgets($s); } }`, 0},
		{"guard does not exit", open + `if (!stream_set_timeout($s, 1)) { fwrite($s, "x"); } return fgets($s); }`, 1},
		{"exit on success", open + `if (stream_set_timeout($s, 1)) { return null; } return fgets($s); }`, 1},
		{"exit after read", open + `$l = fgets($s); if (!stream_set_timeout($s, 1)) { return null; } return $l; }`, 1},
		{"exit in nested list", open + `if ($h) { if (!stream_set_timeout($s, 1)) { return null; } } return fgets($s); }`, 1},
		{"exit on conjunction", open + `if (!$s && !stream_set_timeout($s, 1)) { return null; } return fgets($s); }`, 1},
		{"inverted branch", open + `if (!stream_set_timeout($s, 1)) { return fgets($s); } return null; }`, 1},
		{"failed comparison branch", open + `if (stream_set_timeout($s, 1) === false) { return fgets($s); } }`, 1},
		{"disjunction branch", open + `if ($h || stream_set_timeout($s, 1)) { return fgets($s); } }`, 1},
		{"read in condition", open + `if (stream_set_timeout($s, 1) . fgets($s)) { return 1; } }`, 1},
		{"else branch", open + `if ($s && stream_set_timeout($s, 1)) { return 1; } else { return fgets($s); } }`, 1},
		{"else after failure", open + `if (!stream_set_timeout($s, 1)) { return 1; } else { return fgets($s); } }`, 0},
		{"elseif after failure", open + `if (!stream_set_timeout($s, 1)) { return 1; } elseif ($h) { return fgets($s); } }`, 0},
		{"elseif condition after failure", open + `if (!stream_set_timeout($s, 1)) { return 1; } elseif (fgets($s)) { return 2; } }`, 0},
		{"elseif guard", open + `if (!$s) { return 1; } elseif (stream_set_timeout($s, 1)) { return fgets($s); } }`, 0},
		{"elseif own condition", open + `if (!$s) { return 1; } elseif (fgets($s)) { return 2; } }`, 1},
		{"later elseif", open + `if (!$s) { return 1; } elseif (!stream_set_timeout($s, 1)) { return 2; } elseif ($h) { return fgets($s); } }`, 0},
		{"else after elseif failure", open + `if (!$s) { return 1; } elseif (!stream_set_timeout($s, 1)) { return 2; } else { return fgets($s); } }`, 0},
		{"else without guard", open + `if (!$s) { return 1; } elseif ($h) { return 2; } else { return fgets($s); } }`, 1},
		{"else if", open + `if (!$s) { return 1; } else if (stream_set_timeout($s, 1)) { return fgets($s); } }`, 0},
		{"short circuit and", open + `return stream_set_timeout($s, 1) && fgets($s); }`, 0},
		{"short circuit and keyword", open + `$r = (stream_set_timeout($s, 1) and fgets($s)); return $r; }`, 0},
		{"short circuit or", open + `return !stream_set_timeout($s, 1) || fgets($s); }`, 0},
		{"short circuit or keyword", open + `$r = (!stream_set_timeout($s, 1) or fgets($s)); return $r; }`, 0},
		{"short circuit left", open + `return fgets($s) && stream_set_timeout($s, 1); }`, 1},
		{"unguarded or", open + `return stream_set_timeout($s, 1) || fgets($s); }`, 1},
		{"ternary then", open + `return stream_set_timeout($s, 1) ? fgets($s) : null; }`, 0},
		{"ternary else", open + `return !stream_set_timeout($s, 1) ? null : fgets($s); }`, 0},
		{"ternary wrong arm", open + `return stream_set_timeout($s, 1) ? null : fgets($s); }`, 1},
		{"ternary condition", open + `return fgets($s) ? stream_set_timeout($s, 1) : null; }`, 1},
		{"short ternary", open + `return stream_set_timeout($s, 1) ?: fgets($s); }`, 1},
		{"other handle", open + `$t = fsockopen($h, 81, $errno, $errstr, 2); if (!stream_set_timeout($t, 1)) { return null; } return fgets($s); }`, 1},
		{"rebound handle", open + `if (!stream_set_timeout($s, 1)) { return null; } $s = fsockopen($h, 81, $errno, $errstr, 2); return fgets($s); }`, 1},
		{"jump label after the check", open + `if (!stream_set_timeout($s, 1)) { return; } back: fgets($s); goto back; }`, 1},
		{"unchecked statement", open + `stream_set_timeout($s, 1); return fgets($s); }`, 1},
		{"bool never identical to an int", open + `if (stream_set_timeout($s, 1) === 1) { return fgets($s); } }`, 0},
		{"other operator", open + `if (stream_set_timeout($s, 1) > 0) { return fgets($s); } }`, 1},
		{"other unary", open + `if (-stream_set_timeout($s, 1)) { return fgets($s); } }`, 1},
		{"other call", open + `if (stream_set_blocking($s, true)) { return fgets($s); } }`, 1},
		{"guard outside closure", open + `if (!stream_set_timeout($s, 1)) { return null; } return function () use ($s) { $s = fsockopen("h", 80, $n, $m, 2); return fgets($s); }; }`, 1},
		{"top level early exit", `<?php $s = fsockopen("h", 80, $errno, $errstr, 2); if (!$s || !stream_set_timeout($s, 1)) { exit(1); } echo fgets($s);`, 0},
		{"unbraced body", open + `if ($h) return fgets($s); }`, 1},
		{"float timeout", `<?php function f() { $s = fsockopen("h", 80, $errno, $errstr, 2.5); return fread($s, 10); }`, 1},
		{"float separator", `<?php function f() { $s = fsockopen("h", 80, $errno, $errstr, 1_0.5); return fread($s, 10); }`, 1},
		{"float exponent", `<?php function f() { $s = fsockopen("h", 80, $errno, $errstr, 5e-1); return fread($s, 10); }`, 1},
		{"float zero", `<?php function f() { $s = fsockopen("h", 80, $errno, $errstr, 0.0); return fread($s, 10); }`, 0},
		{"string timeout", `<?php function f() { $s = fsockopen("h", 80, $errno, $errstr, "2"); return fread($s, 10); }`, 0},
		{"dynamic timeout", `<?php function f($t) { $s = fsockopen("h", 80, $errno, $errstr, $t); return fread($s, 10); }`, 0},
		{"named timeout", `<?php function f() { $s = fsockopen("h", 80, timeout: 2); return stream_get_contents($s); }`, 1},
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
