package untrustedshellcommand

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
		{"match constants", "<?php $m = match ($_GET['m']) { 'full' => '--full', default => '--quick' }; system('backup ' . $m);", 0},
		{"match tainted arm", "<?php $m = match ($_GET['m']) { 'raw' => $_GET['v'], default => '--quick' }; system('backup ' . $m);", 1},
		{"match escaped arm", "<?php system('backup ' . match ($_GET['m']) { 'raw' => escapeshellarg($_GET['v']), default => '' });", 0},
		{"length", "<?php system('head -c ' . strlen($_POST['body']) . ' /dev/zero');", 0},
		{"count", "<?php system('seq ' . count($_GET['ids']));", 0},
		{"modulo", "<?php system('sleep ' . ($_GET['n'] % 5));", 0},
		{"sum with literal", "<?php system('sleep ' . ($_GET['n'] + 1));", 0},
		{"array union", "<?php system(implode(' ', $_GET['a'] + $_GET['b']));", 1},
		{"sum of unknowns", "<?php system('sleep ' . ($_GET['a'] + $_GET['b']));", 1},
		{"bitwise strings", "<?php system('tool ' . ($_GET['a'] ^ $_GET['b']));", 1},
		{"text builtin", "<?php system('tool ' . strtolower($_GET['a']));", 1},
		{"user function named like a builtin", "<?php namespace App; function strlen($s) { return $s; } \\system('tool ' . strlen($_GET['a']));", 1},
		{"reassigned after measuring", "<?php $n = strlen($_GET['a']); $n = $_GET['a']; system('tool ' . $n);", 1},
		{"backtick", "<?php $out = `convert {$_GET['file']} out.png`;", 1},
		{"backtick simple variable", "<?php $f = $_GET['file']; `convert $f out.png`;", 1},
		{"backtick escaped", "<?php $f = escapeshellarg($_GET['file']); `convert $f out.png`;", 0},
		{"backtick constant", "<?php `uptime`;", 0},
		{"double-quoted string is not a sink", "<?php $s = \"convert {$_GET['file']}\";", 0},
		{"proc_open string", "<?php proc_open('convert ' . $_GET['file'], [], $pipes);", 1},
		{"proc_open named", "<?php proc_open(descriptor_spec: [], pipes: $pipes, command: $_GET['cmd']);", 1},
		{"proc_open vector", "<?php proc_open(['convert', $_GET['file']], [], $pipes);", 0},
		{"proc_open vector variable", "<?php $argv = ['convert', $_GET['file']]; proc_open($argv, [], $pipes);", 0},
		{"wrapper constants", "<?php function run($c) { system($c); } $m = match ($_GET['m']) { 'a' => '-a', default => '-b' }; run('tool ' . $m);", 0},
		{"wrapper tainted", "<?php function run($c) { system($c); } run('tool ' . $_GET['m']);", 1},
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
