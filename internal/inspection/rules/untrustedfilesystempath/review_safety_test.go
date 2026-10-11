package untrustedfilesystempath

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, Only: []string{"UntrustedFilesystemPath"}})
	if err != nil {
		t.Fatal(err)
	}
	const (
		canonical = "$p = realpath('/srv/d/' . $_GET['f']); "
		checked   = canonical + "if (!str_starts_with($p, '/srv/d/')) { exit; } "
		base      = "$b = realpath('/srv/d'); if ($b === false) { exit; } $p = realpath($b . '/' . $_GET['f']); "
	)
	for _, tc := range []struct {
		name, body string
		findings   int
	}{
		{"literal prefix", checked + "readfile($p);", 0},
		{"parenthesised", canonical + "if ((!(str_starts_with(($p), '/srv/d/')))) { exit; } readfile(($p));", 0},
		{"strpos", canonical + "if (strpos($p, '/srv/d/') !== 0) { return; } readfile($p);", 0},
		{"strpos body", canonical + "if (strpos($p, '/srv/d/') === 0) { readfile($p); }", 0},
		{"longer strncmp", canonical + "if (strncmp($p, '/srv/d/', 20) === 0) { readfile($p); }", 0},
		{"enclosing block", checked + "if ($_GET['raw']) { fopen($p, 'r'); }", 0},
		{"switch case", canonical + "switch ($_GET['m']) { case 'r': if (!str_starts_with($p, '/srv/d/')) { return; } readfile($p); }", 0},
		{"unbraced", canonical + "if (!str_starts_with($p, '/srv/d/')) return; if ($_GET['raw']) readfile($p);", 0},
		{"base slash", base + "if (!str_starts_with($p, $b . '/')) { exit; } readfile($p);", 0},
		{"base separator", base + "if (!str_starts_with($p, $b . DIRECTORY_SEPARATOR)) { exit; } readfile($p);", 0},
		{"value capture", checked + "$f = function () use ($p) { return $p; }; readfile($p);", 0},
		{"other catch", checked + "try { work(); } catch (Exception $e) { } catch (Error) { } readfile($p);", 0},
		{"other global", checked + "global $config; readfile($p);", 0},
		{"contained operand", checked + "readfile($p . '.bak');", 0},
		{"finite match", "$n = match ($_GET['d']) { 'a' => 'a.txt', default => 'b.txt' }; readfile('/srv/d/' . $n);", 0},

		{"no argument", "readfile();", 0},
		{"inline canonical", "readfile(realpath('/srv/d/' . $_GET['f']));", 1},
		{"short strncmp", canonical + "if (strncmp($p, '/srv/d/', 6) !== 0) { exit; } readfile($p);", 1},
		{"dynamic strncmp length", canonical + "if (strncmp($p, '/srv/d/', $_GET['n']) !== 0) { exit; } readfile($p);", 1},
		{"base strncmp", base + "if (strncmp($p, $b . '/', 3) !== 0) { exit; } readfile($p);", 1},
		{"strpos offset", canonical + "if (strpos($p, '/srv/d/', 1) !== 0) { exit; } readfile($p);", 1},
		{"wrong polarity", canonical + "if (strpos($p, '/srv/d/') === 0) { exit; } readfile($p);", 1},
		{"other comparison", canonical + "if (strpos($p, '/srv/d/') === 1) { exit; } readfile($p);", 1},
		{"no call operand", canonical + "if ($p !== 0) { exit; } readfile($p);", 0}, // the flow proves $p is 0
		{"other function", canonical + "if (!str_contains($p, '/srv/d/')) { exit; } readfile($p);", 1},
		{"compared starts_with", canonical + "if (str_starts_with($p, '/srv/d/') !== 0) { exit; } readfile($p);", 1},
		{"user function", canonical + "if (!inside($p, '/srv/d/')) { exit; } readfile($p);", 1}, // an unresolved helper proves nothing, declared or not
		{"other subject", canonical + "$o = realpath('/srv/d/x'); if (!str_starts_with($o, '/srv/d/')) { exit; } readfile($p);", 1},
		{"subject expression", canonical + "if (!str_starts_with($p . '', '/srv/d/')) { exit; } readfile($p);", 1},
		{"root prefix", canonical + "if (!str_starts_with($p, '/')) { exit; } readfile($p);", 1},
		{"relative prefix", canonical + "if (!str_starts_with($p, 'srv/')) { exit; } readfile($p);", 1},
		{"open prefix", canonical + "if (!str_starts_with($p, '/srv/d')) { exit; } readfile($p);", 1},
		{"request prefix", canonical + "if (!str_starts_with($p, $_GET['b'])) { exit; } readfile($p);", 1},
		{"request base", canonical + "if (!str_starts_with($p, $_GET['b'] . '/')) { exit; } readfile($p);", 1},
		{"unchecked base", "$b = realpath('/srv/d'); " + canonical + "if (!str_starts_with($p, $b . '/')) { exit; } readfile($p);", 1},
		{"request-derived base", "$b = realpath($_GET['d']); if ($b === false) { exit; } " + canonical + "if (!str_starts_with($p, $b . '/')) { exit; } readfile($p);", 1},
		{"plain base", "$b = '/srv/d'; " + canonical + "if (!str_starts_with($p, $b . $_GET['s'])) { exit; } readfile($p);", 1},
		{"base suffix", base + "if (!str_starts_with($p, $b . 'x')) { exit; } readfile($p);", 1},
		{"base expression suffix", base + "if (!str_starts_with($p, $b . PHP_EOL)) { exit; } readfile($p);", 1},
		{"base call suffix", base + "if (!str_starts_with($p, $b . sep())) { exit; } readfile($p);", 1},
		{"optional rejection", canonical + "if (!str_starts_with($p, '/srv/d/') && $_GET['strict']) { exit; } readfile($p);", 1},
		{"optional acceptance", canonical + "if (str_starts_with($p, '/srv/d/') || $_GET['any']) { readfile($p); }", 1},
		{"negated call", canonical + "if (-str_starts_with($p, '/srv/d/')) { readfile($p); }", 1},
		{"later guard", canonical + "readfile($p); if (!str_starts_with($p, '/srv/d/')) { exit; }", 1},
		{"loop rebinding", checked + "while ($_GET['more']) { readfile($p); $p = realpath($_GET['next']); }", 1},
		{"extract", checked + "extract($_GET); readfile($p);", 1},
		{"catch binding", checked + "try { work(); } catch (Exception $p) { } readfile($p);", 1},
		{"superseded write", "$p = $_GET['f']; $p = realpath('/srv/d/' . $p); if (!str_starts_with($p, '/srv/d/')) { exit; } readfile($p);", 0},
		{"surplus argument", checked + "readfile($p); preg_match('/x/', 'x', $m, 0, 0, $p);", 0},
		{"unresolved callee", checked + "readfile($p); $_GET['o']->m($p);", 0},
		{"named value argument", checked + "readfile($p); strlen(string: $p);", 0},
		{"later compound write", checked + "readfile($p); $p .= 'x';", 1},
		{"later destructuring", checked + "readfile($p); [$p] = [1];", 1},
		{"later list slot", checked + "readfile($p); list(, $p) = $_GET['l'];", 1},
		{"later offset write", checked + "readfile($p); $p[0] = 'x';", 1},
		{"later increment", checked + "readfile($p); $p++;", 1},
		{"later loop value", checked + "readfile($p); foreach ([] as $p) { }", 1},
		{"later loop key", checked + "readfile($p); foreach ([] as $p => $v) { }", 1},
		{"reference alias", checked + "readfile($p); $a = &$p;", 1},
		{"variable variable", checked + "readfile($p); $$n = 1;", 1},
		{"named output argument", checked + "readfile($p); settype(var: $p, type: 'string');", 1},
		{"variadic output argument", checked + "readfile($p); sscanf('a b', '%s %s', $x, $p);", 1},
		{"static declaration", checked + "readfile($p); static $p;", 1},
		{"capture beside a call", checked + "$x = (function () use (&$p) { })() . strlen('a'); readfile($p);", 1},
		{"base rebound", base + "if (!str_starts_with($p, $b . '/')) { exit; } readfile($p); $b .= 'x';", 1},
		{"inline match", "readfile('/srv/d/' . match ($_GET['d']) { 'a' => 'a.txt', default => 'b.txt' });", 0},
		{"rebound match", "$n = match ($_GET['d']) { 'a' => 'a.txt', default => 'b.txt' }; extract($_GET); readfile('/srv/d/' . $n);", 1},
		{"request arm", "$n = match ($_GET['d']) { 'a' => 'a.txt', default => $_GET['d'] }; readfile('/srv/d/' . $n);", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, source := range []string{"<?php function probe() { " + tc.body + " }", "<?php " + tc.body} {
				f := syntax.ParseBest("review.php", []byte(source), syntax.Options{Version: phpversion.PHP85})
				if len(f.Errors) != 0 {
					t.Fatalf("invalid probe: %v", f.Errors)
				}
				if got := e.Analyze(f); len(got) != tc.findings {
					t.Fatalf("%s\nfindings: got %v, want %d", source, got, tc.findings)
				}
			}
		})
	}
}

func TestReviewGlobalRebinding(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, Only: []string{"UntrustedFilesystemPath"}})
	if err != nil {
		t.Fatal(err)
	}
	const guarded = "$p = realpath('/srv/d/' . $_GET['f']); if (!str_starts_with($p, '/srv/d/')) { exit; } "
	for source, findings := range map[string]int{
		// The called function writes request data to the global.
		"<?php function swap() { global $p; $p = $_GET['g']; } " + guarded + "swap(); readfile($p);": 1,
		// A declaring function that is never called leaves the rebinding to
		// the local containment proof, which refuses it.
		"<?php function swap() { global $p; $p = $_GET['g']; } " + guarded + "readfile($p);": 1,
		"<?php " + guarded + "$all = $GLOBALS; readfile($p);":                                1,
	} {
		f := syntax.ParseBest("review.php", []byte(source), syntax.Options{Version: phpversion.PHP85})
		if got := e.Analyze(f); len(f.Errors) != 0 || len(got) != findings {
			t.Fatalf("%s\nerrors %v, findings %v", source, f.Errors, got)
		}
	}
}
