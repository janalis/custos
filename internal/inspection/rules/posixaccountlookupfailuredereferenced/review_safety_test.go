package posixaccountlookupfailuredereferenced

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewGuards(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	const head = "<?php function f($u, $o) { $i = posix_getpwnam($u); "
	for _, tc := range []struct {
		name, body string
		findings   int
	}{
		{"strict ternary", "return $i !== false ? $i['dir'] : null;", 0},
		{"strict ternary else", "return $i === false ? null : $i['dir'];", 0},
		{"truthy ternary", "return $i ? $i['dir'] : null;", 0},
		{"array ternary", "return is_array($i) ? $i['dir'] : null;", 0},
		{"strict or chain", "if ($i === false || $i['uid'] === 0) { return null; } return $i['dir'];", 0},
		{"and chain", "return $i !== false && $i['uid'] === 0;", 0},
		{"keyword and chain", "return ($i and $i['uid']);", 0},
		{"keyword or chain", "return (!$i or $i['uid']);", 0},
		{"isset chain", "return isset($i['dir']) && $i['dir'] !== '';", 0},
		{"falsy exit", "if (!$i) { return null; } return $i['dir'];", 0},
		{"not array exit", "if (!\\is_array($i)) { throw new Exception(); } return $i['dir'];", 0},
		{"array branch", "if (is_array($i)) { return $i['dir']; } return null;", 0},
		{"isset branch", "if (isset($i['a']['b'])) { return $i['dir']; } return null;", 0},
		{"isset exit", "if (!isset($i['uid'], $i['dir'])) { return; } echo $i['uid'] . $i['dir'];", 0},
		{"not empty key", "if (!empty($i['dir'])) { echo $i['dir']; }", 0},
		{"empty exit", "if (empty($i)) { return; } echo $i['dir'];", 0},
		{"loose exit", "if ($i == false) { return; } echo $i['dir'];", 0},
		{"loose yoda branch", "if (\\FALSE != $i) { echo $i['dir']; }", 0},
		{"and branch", "if ($o && $i) { echo $i['dir']; }", 0},
		{"or exit", "if (!is_array($i) || !isset($i['dir'])) { return; } echo $i['dir'];", 0},
		{"negated or branch", "if (!($i === false || $o)) { echo $i['dir']; }", 0},
		{"and exit", "if (!$i && empty($i)) { return; } echo $i['dir'];", 0},
		{"or branch", "if (is_array($i) || isset($i['x'])) { echo $i['dir']; }", 0},
		{"else branch", "if (!$i) { echo 0; } else { echo $i['dir']; }", 0},
		{"elseif branch", "if ($o) { echo 0; } elseif ($i) { echo $i['dir']; }", 0},
		{"after failed elseif", "if ($o) { echo 0; } elseif (!$i) { echo 1; } else { echo $i['dir']; }", 0},
		{"elseif after failed if", "if (!$i) { echo 0; } elseif ($o) { echo $i['dir']; }", 0},
		{"outer exit", "if (!$i) { return; } echo 1; if ($o) { echo $i['dir']; }", 0},
		{"or throw", "$i !== false or throw new Exception(); return $i['dir'];", 0},
		{"or exit statement", "$i || exit(1); return $i['dir'];", 0},
		{"and die", "!$i and die('unknown'); return $i['dir'];", 0},
		{"assigned or throw", "$j = posix_getpwuid($o) or throw new Exception(); return $j['dir'];", 0},
		{"assigned exit", "if (!($j = posix_getpwuid($o))) { return; } return $j['dir'];", 0},
		// Key probes are proven by the guard walk only.
		{"isset ternary", "return isset($i['dir']) ? $i['dir'] : null;", 0},
		{"isset else branch", "if (!isset($i['dir'])) { echo 0; } else { echo $i['dir']; }", 0},
		{"after failed isset elseif", "if ($o) { echo 0; } elseif (!isset($i['dir'])) { echo 1; } else { echo $i['dir']; }", 0},
		{"isset and branch", "if ($o && isset($i['dir'])) { echo $i['dir']; }", 0},
		{"isset or exit", "if (!isset($i['dir']) || $o) { return; } echo $i['dir'];", 0},

		{"unguarded", "return $i['dir'];", 1},
		{"goto lands after guard", "if (!$i) { goto x; } echo 1; x: return $i['dir'];", 1},
		{"unrelated truthy", "if ($o) { echo $i['dir']; }", 1},
		{"unbraced unrelated", "if ($o) echo $i['dir'];", 1},
		{"wrong branch", "if (!$i) { echo $i['dir']; }", 1},
		{"wrong ternary arm", "return $i ? null : $i['dir'];", 1},
		{"short ternary", "return $o ?: $i['dir'];", 1},
		{"plain isset", "if (isset($i)) { echo $i['dir']; }", 1},
		{"negated isset branch", "if (!isset($i['dir'])) { echo $i['dir']; }", 1},
		{"empty branch", "if (empty($i)) { echo $i['dir']; }", 1},
		{"negated array branch", "if (!is_array($i)) { echo $i['dir']; }", 1},
		{"other array", "if (is_array($o)) { echo $i['dir']; }", 1},
		{"other call", "if (strlen($i)) { echo $i['dir']; }", 1},
		{"null comparison", "if ($i !== null) { echo $i['dir']; }", 1},
		{"true comparison", "if ($i === true) { echo $i['dir']; }", 1},
		{"identical branch", "if ($i === false) { echo $i['dir']; }", 1},
		{"not identical exit", "if ($i !== false) { return; } echo $i['dir'];", 1},
		{"relational", "if ($i > 0) { echo $i['dir']; }", 1},
		{"negated relational", "if (!($i > 0)) { echo $i['dir']; }", 1},
		{"minus", "if (-$i) { echo $i['dir']; }", 1},
		{"eager operator", "return !$i . $i['dir'];", 1},
		{"left operand", "return $i['dir'] && $i;", 1},
		{"or keeps failure", "return $i || $i['dir'];", 1},
		{"and keeps failure", "return !$i && $i['dir'];", 1},
		{"condition read", "if ($i['dir']) { echo 1; }", 1},
		{"elseif condition read", "if ($o) { echo 0; } elseif ($i['dir']) { echo 1; }", 1},
		{"non exiting branch", "if (!$i) { echo 0; } echo $i['dir'];", 1},
		{"unrelated else", "if ($o) { echo 0; } else { echo $i['dir']; }", 1},
		{"unrelated statement", "echo $o; return $i['dir'];", 1},
		{"later exit", "echo $i['dir']; if (!$i) { return; }", 1},
		{"plain statement", "$i !== false; return $i['dir'];", 1},
		{"eager statement", "$i . exit(1); return $i['dir'];", 1},
		{"or without abort", "$i or print('x'); return $i['dir'];", 1},
		{"wrong abort", "$i and exit(1); return $i['dir'];", 1},
		{"other assigned abort", "$j = $o or throw new Exception(); return $i['dir'];", 1},
		{"compound assigned abort", "$j = posix_getpwuid($o); $j[0] = posix_getpwuid($o) or exit; return $i['dir'];", 1},
		{"test before lookup", "if (!$i) { return; } $i = posix_getpwnam($o); return $i['dir'];", 1},
		{"reassigned after test", "if (!$i) { return; } $i = posix_getpwnam($o); return $i['dir'];", 1},
		{"direct call", "if (posix_getpwuid(5)) { echo posix_getpwuid(5)['name']; }", 1},
	} {
		// A scope the flow analysis does not cover (here through a trailing
		// `goto` beside a `try`) is judged by the guards written around the read alone.
		for _, tail := range []string{" }", " try { } finally { } goto done; done: ; }"} {
			t.Run(tc.name+tail, func(t *testing.T) {
				f := syntax.ParseBest("review.php", []byte(head+tc.body+tail), syntax.Options{Version: phpversion.PHP85})
				if len(f.Errors) != 0 {
					t.Fatalf("invalid probe: %v", f.Errors)
				}
				if got := e.Analyze(f); len(got) != tc.findings {
					t.Fatalf("findings: got %v, want %d", got, tc.findings)
				}
			})
		}
	}
}

// Several lookups can reach one read: each needs a test on its own path.
func TestReviewReassignedLookups(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, body string
		findings   int
	}{
		// The earlier test does not vouch for the lookup made on one path.
		{"reassigned on one path after test", "if (!$i) { return; } if ($o) { $i = posix_getpwnam($o); } return $i['dir'];", 1},
		{"reassigned on one path and tested again", "if (!$i) { return; } if ($o) { $i = posix_getpwnam($o); } if (!$i) { return; } return $i['dir'];", 0},
		{"reassigned in a loop after test", "if (!$i) { return; } while ($o--) { echo $i['dir']; $i = posix_getpwnam($o); }", 1},
		{"tested in the loop that reassigns", "while ($o--) { if (!$i) { return; } echo $i['dir']; $i = posix_getpwnam($o); }", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("review.php", []byte("<?php function f($u, $o) { $i = posix_getpwnam($u); "+tc.body+" }"), syntax.Options{Version: phpversion.PHP85})
			if got := e.Analyze(f); len(f.Errors) != 0 || len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d (%v)", got, tc.findings, f.Errors)
			}
		})
	}
	// The unguarded result of another call reaching the same read does not
	// hide a lookup, and does not stand in for one.
	for _, tc := range []struct {
		name, body string
		findings   int
	}{
		{"other call first, lookup on one path", "$i = scandir($u); if ($o) { $i = posix_getpwnam($u); } return $i['dir'];", 1},
		{"other call first and checked, lookup on one path", "$i = scandir($u); if ($i === false) { return; } if ($o) { $i = posix_getpwnam($u); } return $i['dir'];", 1},
		{"text first, lookup in a loop", "$i = trim($u); while ($o--) { $i = posix_getpwuid($o); } return $i['dir'];", 1},
		{"other call first, lookup checked on its path", "$i = scandir($u); if ($o) { $i = posix_getpwnam($u); if (!$i) { return; } } return $i['dir'];", 0},
		{"lookup checked, other call on one path", "$i = posix_getpwnam($u); if ($i === false) { return; } if ($o) { $i = scandir($u); } return $i['dir'];", 0},
		{"two other calls", "$i = scandir($u); if ($o) { $i = glob($u); } return $i['dir'];", 0},
		{"other call first, both checked after", "$i = scandir($u); if ($o) { $i = posix_getpwnam($u); } if (!$i) { return; } return $i['dir'];", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("review.php", []byte("<?php function f($u, $o) { "+tc.body+" }"), syntax.Options{Version: phpversion.PHP85})
			if got := e.Analyze(f); len(f.Errors) != 0 || len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d (%v)", got, tc.findings, f.Errors)
			}
		})
	}
}
