package untrustednetworkdestination

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
		{"qualified option", "<?php $ch = curl_init(); curl_setopt($ch, \\CURLOPT_URL, $_GET['u']); curl_exec($ch);", 1},
		{"namespaced calls", "<?php namespace App; $ch = \\curl_init(); \\curl_setopt($ch, \\CURLOPT_URL, $_GET['u']); \\curl_exec($ch);", 1},
		{"namespace fallback", "<?php namespace App; $ch = curl_init(); curl_setopt($ch, CURLOPT_URL, $_GET['u']); curl_exec($ch);", 1},
		{"namespaced constant", "<?php namespace App; const CURLOPT_URL = 1; $ch = \\curl_init(); \\curl_setopt($ch, CURLOPT_URL, $_GET['u']); \\curl_exec($ch);", 0},
		{"option array", "<?php $ch = curl_init(); curl_setopt_array($ch, [CURLOPT_URL => $_GET['u'], CURLOPT_RETURNTRANSFER => true]); curl_exec($ch);", 1},
		{"option array variable", "<?php $ch = curl_init(); $o = [\\CURLOPT_URL => $_GET['u']]; curl_setopt_array($ch, $o); curl_exec($ch);", 1},
		{"option array replaced", "<?php $ch = curl_init(); curl_setopt_array($ch, [CURLOPT_URL => $_GET['u']]); curl_setopt($ch, CURLOPT_URL, 'https://example.com/'); curl_exec($ch);", 0},
		{"option replaced by array", "<?php $ch = curl_init(); curl_setopt($ch, CURLOPT_URL, $_GET['u']); curl_setopt_array($ch, [CURLOPT_URL => 'https://example.com/']); curl_exec($ch);", 0},
		{"stream read", "<?php echo strlen(file_get_contents($_GET['u']));", 1},
		{"stream named argument", "<?php readfile(filename: $_GET['u']);", 1},
		{"stream request host", "<?php $h = fopen('http://' . $_GET['host'] . '/status', 'r');", 1},
		{"stream ftp host", "<?php copy('FTP://' . $_GET['host'] . '/dump', '/tmp/dump');", 1},
		{"stream fixed host", "<?php readfile('https://cdn.example.com/' . $_GET['name']);", 0},
		{"stream local prefix", "<?php readfile('/srv/files/' . $_GET['name']);", 0},
		{"stream local wrapper", "<?php readfile('php://' . $_GET['name']);", 0},
		{"stream numeric lead", "<?php readfile(intval($_GET['id']) . '.csv');", 0},
		{"stream copy target", "<?php copy('/srv/files/a', $_GET['to']);", 0},
		{"stream guarded", "<?php $u = $_GET['u']; if (!in_array($u, ['https://a.example/', 'https://b.example/'], true)) { return; } readfile($u);", 0},
		{"user function", "<?php namespace App; function readfile($f) {} readfile($_GET['u']);", 0},
		{"fixed host query", "<?php $ch = curl_init('https://api.example.com/search?q=' . urlencode($_GET['q'])); curl_exec($ch);", 0},
		{"fixed host no path", "<?php $ch = curl_init('https://api.example.com?q=' . $_GET['q']); curl_exec($ch);", 0},
		{"fixed host through variables", "<?php $base = 'https://api.example.com/'; $url = $base . 'items/' . $_GET['id']; $ch = curl_init(($url)); curl_exec($ch);", 0},
		{"open host", "<?php $ch = curl_init('https://api.example.com' . $_GET['q']); curl_exec($ch);", 1},
		{"open credentials", "<?php $ch = curl_init('https://user@' . $_GET['q'] . '/'); curl_exec($ch);", 1},
		{"request scheme", "<?php $ch = curl_init($_GET['scheme'] . '://api.example.com/'); curl_exec($ch);", 1},
		{"empty authority", "<?php $ch = curl_init('file:///' . $_GET['path']); curl_exec($ch);", 1},
		{"heredoc prefix", "<?php $ch = curl_init(<<<'U'\nhttps://api.example.com/\nU . $_GET['q']); curl_exec($ch);", 1},
		{"match literals", "<?php $dir = match ($_GET['env']) { 'eu' => 'https://eu.example.com/', default => 'https://us.example.com/' }; $c = curl_init($dir); curl_exec($c);", 0},
		{"match inline", "<?php $c = curl_init(match ($_GET['env']) { 'eu' => 'https://eu.example.com/', default => 'https://us.example.com/' }); curl_exec($c);", 0},
		{"match variable arm", "<?php $safe = 'https://eu.example.com/'; $c = curl_init(match ($_GET['env']) { 'eu' => $safe, default => 'https://us.example.com/' }); curl_exec($c);", 1},
		{"match request arm", "<?php $c = curl_init(match ($_GET['env']) { 'eu' => $_GET['u'], default => 'https://us.example.com/' }); curl_exec($c);", 1},
		{"match with suffix", "<?php $c = curl_init(match ($_GET['env']) { default => 'https://us.example.com' } . $_GET['x']); curl_exec($c);", 1},
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
