package unvalidatedredirecttarget

import (
	"strings"
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
	literalSelection := func(depth int) string {
		return strings.Repeat("match ($_GET['k']) { default => ", depth) + "'/home'" + strings.Repeat(" }", depth)
	}
	for _, tc := range []struct {
		name, source string
		findings     int
	}{
		{"direct", "<?php header('Location: ' . $_GET['u']);", 1},
		{"suffix", "<?php header('Location: ' . $_GET['u'] . '?saved=1');", 1},
		{"path suffix", "<?php header('Location: /items/' . $_GET['id'] . '/edit');", 1},
		{"interpolated", "<?php header(\"Location: {$_GET['u']}\");", 1},
		{"simple interpolation", "<?php $u = $_GET['u']; header(\"Location: $u\");", 1},
		{"heredoc", "<?php header(<<<TXT\nLocation: {$_GET['u']}\nTXT);", 1},
		{"lowercase name", "<?php header('location: ' . $_GET['u'] . '#top');", 1},
		{"named argument", "<?php header(replace: true, header: 'Location: ' . $_GET['u'] . '/');", 1},
		{"constant name", "<?php const NAME = 'Location: '; header(NAME . $_GET['u'] . '/');", 1},
		{"other header", "<?php header('Refresh: 0; url=' . $_GET['u'] . '&x=1');", 0},
		{"literal target", "<?php header('Location: ' . '/home' . '?saved=1');", 0},
		{"unknown value", "<?php function f($u) { header('Location: ' . $u . '/x'); }", 0},
		{"allowlisted", "<?php $u = $_GET['u']; if (!in_array($u, ['/a', '/b'], true)) { exit; } header(\"Location: {$u}\");", 0},
		{"literal match", "<?php header('Location: ' . match ($_GET['k']) { 'a' => '/a', default => '/b' });", 0},
		{"literal match variable", "<?php $p = match ($_GET['k']) { 'a' => '/a', default => '/b' }; header('Location: ' . $p);", 0},
		{"nested literal match", "<?php header('Location: ' . " + literalSelection(maxSelectionDepth) + ");", 0},
		{"selection budget", "<?php header('Location: ' . " + literalSelection(maxSelectionDepth+1) + ");", 1},
		{"match returning input", "<?php header('Location: ' . match ($_GET['k']) { 'a' => '/a', default => $_GET['u'] });", 1},
		{"match building input", "<?php $p = match ($_GET['k']) { 'a' => '/a', default => '/b/' . $_GET['u'] }; header('Location: ' . $p);", 1},
		{"ternary input", "<?php $p = isset($_GET['u']) ? $_GET['u'] : '/a'; header('Location: ' . $p);", 1},
		{"encoded query", "<?php header('Location: /search?q=' . urlencode($_GET['q']));", 0},
		{"raw encoded query", "<?php header('Location:/search?q=' . rawurlencode($_GET['q']) . '&p=1');", 0},
		{"built query", "<?php header('Location: /search?' . http_build_query(['q' => $_GET['q']]));", 0},
		{"encoded query variable", "<?php $q = urlencode($_GET['q']); header(\"Location: /search?q=$q\");", 0},
		{"encoded without path", "<?php header('Location: ' . urlencode($_GET['q']));", 1},
		{"encoded path segment", "<?php header('Location: /items/' . urlencode($_GET['q']));", 1},
		{"encoded on other host", "<?php header('Location: //search?q=' . urlencode($_GET['q']));", 1},
		{"encoded after backslash", "<?php header('Location: /\\\\search?q=' . urlencode($_GET['q']));", 1},
		{"raw query value", "<?php header('Location: /search?q=' . $_GET['q']);", 1},
		{"other call in query", "<?php header('Location: /search?q=' . trim($_GET['q']));", 1},
		{"mixed query values", "<?php header('Location: /search?q=' . urlencode($_GET['q']) . '&r=' . $_GET['r']);", 1},
		{"user encoder", "<?php namespace App; function urlencode($v) { return $v; } header('Location: /search?q=' . urlencode($_GET['q']));", 1},
		{"backtick", "<?php header(`printf 'Location: ' {$_GET['u']}`);", 0},
		{"unpacked", "<?php header(...$_GET['h']);", 0},
		{"whole value", "<?php header($_GET['h']);", 0},
		{"user header", "<?php namespace App; function header($v) {} header('Location: ' . $_GET['u'] . '/');", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("review.php", []byte(tc.source), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
