package unescapedhtmloutput

import (
	"testing"

	"custos/internal/fixing"
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
		{"concatenation", "<?php echo '<p>' . $_GET['v'] . '</p>';", 1},
		// A matched switch label is text of the program: what follows it in
		// the statement is still looked at.
		{"matched element before request data", "<?php function f(){ switch ($_GET['k']) { case 'a': echo '<p>' . $_GET['k'] . $_POST['k'] . '</p>'; } }", 1},
		{"matched element before another key", "<?php function f(){ switch ($_GET['k']) { case 'a': echo '<p>' . $_GET['k'] . $_GET['j'] . '</p>'; } }", 1},
		{"matched variable before request data", "<?php function f(){ $k = $_GET['k']; switch ($k) { case 'a': echo '<p>' . $k . $_POST['k'] . '</p>'; } }", 1},
		{"matched element alone", "<?php function f(){ switch ($_GET['k']) { case 'a': echo '<p>' . $_GET['k'] . '</p>'; } }", 0},
		{"matched element before escaped data", "<?php function f(){ switch ($_GET['k']) { case 'a': echo '<p>' . $_GET['k'] . htmlspecialchars($_POST['k']) . '</p>'; } }", 0},
		{"interpolation", `<?php echo "<p>{$_GET['v']}</p>";`, 1},
		{"simple interpolation", `<?php echo "<p>$_GET[v]</p>";`, 1},
		{"operand list", "<?php echo '<p>', $_GET['v'], '</p>';", 1},
		{"print", "<?php print '<p>' . $_GET['v'] . '</p>';", 1},
		{"after closing tag", "<?php echo '<p>a</p>' . $_GET['v'];", 1},
		{"escaped prefix quote", `<?php echo "<p title=\"a\">{$_GET['v']}";`, 1},
		{"tag reopened", "<?php echo '<p <b>' . $_GET['v'];", 0},
		{"comment prefix", "<?php echo '<!-- c -->' . '<p>' . $_GET['v'];", 0},
		{"dangling bracket", "<?php echo '<p><' . $_GET['v'];", 0},
		{"short echo", "<p><?= $_GET['v'] ?></p>", 0},
		{"match constants", "<?php $d = match ($_GET['s']) { 'desc' => 'DESC', default => 'ASC' }; echo '<p>' . $d;", 0},
		{"match request arm", "<?php $d = match (true) { default => $_GET['s'] }; echo '<p>' . $d;", 1},
		{"arithmetic", "<?php echo '<p>' . ($_GET['a'] + $_GET['b']) . ($_GET['a'] % 2) . ($_GET['a'] ** 2) . ($_GET['a'] / 2);", 0},
		{"array union joined", "<?php echo '<p>' . implode(',', $_GET['a'] + $_GET['b']);", 1},
		{"byte-wise operator", "<?php echo '<p>' . ($_GET['a'] | 'x');", 1},
		{"numeric format", "<?php echo '<p>' . sprintf('%d', $_GET['a']);", 0},
		{"padded format", "<?php echo '<p>' . sprintf(\"%'<5d\", $_GET['a']);", 1},
		{"dangling format", "<?php echo '<p>' . sprintf('%d %', $_GET['a']);", 1},
		{"dynamic format", "<?php echo '<p>' . sprintf($f, $_GET['a']);", 1},
		{"shadowed strlen", "<?php namespace App; function strlen($s) { return $s; } echo '<p>' . strlen($_GET['a']);", 1},
		{"deep selection", "<?php echo '<p>' . ($a ? 1 : ($a ? 1 : ($a ? 1 : ($a ? 1 : ($a ? 1 : ($a ? 1 : ($a ? 1 : ($a ? 1 : ($a ? 1 : ($_GET['v'] + 1))))))))));", 1},
		{"shallow selection", "<?php echo '<p>' . ($a ? 1 : ($a ? 1 : ($_GET['v'] + 1)));", 0},
		{"element overwritten", "<?php $_GET['v'] = 'safe'; echo '<p>' . $_GET['v'];", 0},
		{"other element written", "<?php $_GET['page'] = 1; echo '<p>' . $_GET['v'];", 1},
		{"other element removed", "<?php unset($_GET['page']); echo '<p>' . $_GET['v'];", 1},
		{"element trimmed in place", "<?php $_GET['v'] = trim($_GET['v']); echo '<p>' . $_GET['v'];", 1},
		{"element appended", "<?php $_GET['v'] .= '!'; echo '<p>' . $_GET['v'];", 1},
		{"element escaped in place", "<?php $_GET['v'] = htmlspecialchars($_GET['v']); echo '<p>' . $_GET['v'];", 0},
		{"numeric element overwritten", "<?php $_GET['0'] = 'safe'; echo '<p>' . $_GET[0];", 0},
		{"dynamic element written", "<?php $_GET[$k] = 'safe'; echo '<p>' . $_GET['v'];", 0},
		{"nested element written", "<?php $_GET['v']['x'] = 'safe'; echo '<p>' . $_GET['v'];", 0},
		{"dynamic element read", "<?php $_GET['page'] = 1; echo '<p>' . $_GET[$k];", 0},
		{"whole array read", "<?php $_GET['page'] = 1; echo '<p>' . implode(',', $_GET);", 0},
		{"whole array replaced", "<?php $_GET = array_map('htmlspecialchars', $_GET); echo '<p>' . $_GET['v'];", 0},
		// A reference replaces nothing; only a write through it does.
		{"reference taken", "<?php $r = &$_GET['v']; echo '<p>' . $_GET['v'];", 1},
		{"clean value written through a reference", "<?php $r = &$_GET['v']; $r = 'safe'; echo '<p>' . $_GET['v'];", 0},
		{"escaped through a reference", "<?php $r = &$_GET['v']; $r = htmlspecialchars($r); echo '<p>' . $_GET['v'];", 0},
		{"converted through a reference", "<?php $r = &$_GET['v']; settype($r, 'int'); echo '<p>' . $_GET['v'];", 0},
		{"request written through a reference", "<?php $r = &$_GET['v']; $r = $_POST['x']; echo '<p>' . $_GET['v'];", 1},
		{"element of a reference written", "<?php $r = &$_GET['v']; $r['v'] = 'safe'; echo '<p>' . $_GET['v'];", 0},
		{"clean value written through a second reference", "<?php $r = &$_GET; $s = &$r; $s['v'] = 'safe'; echo '<p>' . $_GET['v'];", 0},
		{"bound to a clean local", "<?php $_GET['v'] = &$clean; echo '<p>' . $_GET['v'];", 0},
		{"bound local cleaned later", "<?php $t = $_POST['x']; $_GET['v'] = &$t; $t = 'safe'; echo '<p>' . $_GET['v'];", 0},
		{"reference held by a property", "<?php $o->p = &$_GET['v']; $o->q = 'safe'; echo '<p>' . $_GET['v'];", 1},
		{"reference between superglobals", "<?php $_POST = &$_GET; $_POST['v'] = 'safe'; echo '<p>' . $_GET['v'];", 0},
		{"unrelated reference", "<?php $r = &$other; $r = 'safe'; echo '<p>' . $_GET['v'];", 1},
		// A write that stores request text keeps the superglobal request-derived.
		{"whole array from request", "<?php $_GET = $_POST; echo '<p>' . $_GET['v'];", 1},
		{"whole array merged", "<?php $_GET = array_merge($_GET, $_POST); echo '<p>' . $_GET['v'];", 1},
		{"whole array united", "<?php $_GET += $_POST; echo '<p>' . $_GET['v'];", 1},
		{"whole array cleared", "<?php $_GET = []; echo '<p>' . $_GET['v'];", 0},
		{"whole array cast", "<?php $_GET = array_map('intval', $_GET); echo '<p>' . $_GET['v'];", 0},
		{"whole array from project code", "<?php $_GET = sanitise($_GET); echo '<p>' . $_GET['v'];", 0},
		{"dynamic element from request", "<?php $_GET[$k] = $_POST['x']; echo '<p>' . $_GET['v'];", 1},
		{"nested element from request", "<?php $_GET['v']['x'] = $_POST['x']; echo '<p>' . $_GET['v'];", 1},
		{"globals entry from request", "<?php $GLOBALS['_GET'] = $_POST; echo '<p>' . $_GET['v'];", 1},
		{"filled from request", "<?php parse_str($_SERVER['QUERY_STRING'], $_GET); echo '<p>' . $_GET['v'];", 1},
		{"filled from a literal", "<?php parse_str('v=1', $_GET); echo '<p>' . $_GET['v'];", 0},
		{"filled by project code", "<?php function fill(array &$to, array $from) { $to = array_map('htmlspecialchars', $from); } fill($_GET, $_POST); echo '<p>' . $_GET['v'];", 0},
		{"walked with a callback", "<?php array_walk($_GET, function (&$v) { $v = htmlspecialchars($v); }); echo '<p>' . $_GET['v'];", 0},
		{"read in a callee", "<?php function q() { return $_GET['v']; } $_GET['page'] = 1; echo '<p>' . q();", 0},
		{"other element then variable", "<?php $_GET['page'] = 1; $v = $_GET['v']; echo '<p>' . $v;", 1},
		{"extracted request", "<?php extract($_GET); end($_GET); echo '<p>' . $_GET['v'];", 1},
		{"origin after extract", "<?php $n = ($_GET['a'] + 1); extract($_GET); echo '<p>' . $n;", 1},
		{"origin captured by reference", "<?php $n = ($_GET['a'] + 1); $f = function () use (&$n) { $n = $_GET['b']; }; $f(); echo '<p>' . $n;", 1},
		{"origin captured beside a call", "<?php $n = ($_GET['a'] + 1); $x = (function () use (&$n) { $n = $_GET['b']; })() . strlen('a'); echo '<p>' . $n;", 1},
		{"origin with alias", "<?php $n = ($_GET['a'] + 1); echo '<p>' . $n; $m = &$n;", 1},
		{"origin iterated by reference", "<?php $n = [($_GET['a'] + 1)]; foreach ($n as &$m) { } echo '<p>' . $n;", 1},
		{"origin bound by catch", "<?php $n = ($_GET['a'] + 1); try { run(); } catch (Exception $n) { } echo '<p>' . $n;", 1},
		{"origin declared static", "<?php function probe() { $n = ($_GET['a'] + 1); echo '<p>' . $n; static $n; }", 1},
		{"origin passed by reference", "<?php $n = ($_GET['a'] + 1); echo '<p>' . $n; settype($n, 'string');", 1},
		{"origin beside variable variable", "<?php $n = ($_GET['a'] + 1); echo '<p>' . $n; $$k = $_GET['b'];", 1},
		{"origin declared global", "<?php function probe() { global $n; $n = ($_GET['a'] + 1); echo '<p>' . $n; }", 1},
		{"origin captured by reference, never invoked", "<?php $n = ($_GET['a'] + 1); $f = function () use (&$n) { $n = $_GET['b']; }; echo '<p>' . $n;", 1},
		{"origin iterated by reference afterwards", "<?php $n = ($_GET['a'] + 1); echo '<p>' . $n; foreach ($n as &$m) { }", 1},
		{"match numeric sum arm", "<?php $d = match ($_GET['s']) { 'a' => $_GET['n'] + 1, default => 0 }; echo '<p>' . $d;", 0},
		{"origin with other catch", "<?php $n = ($_GET['a'] + 1); try { run(); } catch (Exception $e) { } catch (Error) { } echo '<p>' . $n;", 0},
		{"origin beside unrelated scopes", "<?php function other() { extract($_POST); static $n; global $m; } $n = ($_GET['a'] + 1); echo '<p>' . $n;", 0},
		{"origin written through global", "<?php function swap() { global $n; $n = $_GET['b']; } $n = ($_GET['a'] + 1); swap(); echo '<p>' . $n;", 1},
		{"origin beside globals table", "<?php $n = ($_GET['a'] + 1); $all = $GLOBALS; echo '<p>' . $n;", 1},
		{"origin written through globals table", "<?php function swap() { $GLOBALS['n'] = $_GET['b']; } $n = ($_GET['a'] + 1); swap(); echo '<p>' . $n;", 1},
		{"origin beside other global entry", "<?php function swap() { $GLOBALS['m'] = $_GET['b']; } $n = ($_GET['a'] + 1); swap(); echo '<p>' . $n;", 0},
		{"origin beside variable variable read", "<?php $n = ($_GET['a'] + 1); echo '<p>' . $n; $m = $$k;", 0},
		{"negated number", "<?php echo '<p>' . -$_GET['a'] . +$_GET['b'];", 0},
		{"bitwise not", "<?php echo '<p>' . ~$_GET['a'];", 1},
		{"other superglobal written", "<?php $_POST['v'] = 'safe'; echo '<p>' . $_GET['v'];", 1},
		{"written afterwards", "<?php echo '<p>' . $_GET['v']; $_GET['v'] = 'safe';", 1},
		{"variable after overwrite", "<?php $_GET['v'] = 'safe'; $v = $_GET['v']; echo '<p>' . $v;", 0},
		{"sorted in place", "<?php sort($_GET); echo '<p>' . $_GET[0];", 0},
		{"read-only call", "<?php count($_GET); echo '<p>' . $_GET['v'];", 1},
		{"svg response", "<?php header('Content-Type: image/svg+xml'); echo '<p>' . $_GET['v'];", 1},
		{"xhtml response", "<?php header('Content-Type: application/xhtml+xml'); echo '<p>' . $_GET['v'];", 1},
		{"json response", "<?php header('Content-Type: application/json'); echo '<p>' . $_GET['v'];", 0},
		{"csv response", "<?php header('Content-Type: text/csv; charset=UTF-8'); echo '<p>' . $_GET['v'];", 0},
		{"html after plain", "<?php header('Content-Type: text/plain'); header('Content-Type: text/html'); echo '<p>' . $_GET['v'];", 1},
		{"header afterwards", "<?php echo '<p>' . $_GET['v']; header('Content-Type: text/plain');", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte(tc.source)
			opt := syntax.Options{Version: phpversion.PHP85}
			f := syntax.ParseBest("review.php", src, opt)
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			got := e.Analyze(f)
			if len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
			// The rule never rewrites code: no proof of charset or scalar input.
			res := fixing.FixSource(e, "review.php", src, fixing.Options{Parse: opt, SinglePass: true})
			if string(res.Source) != tc.source {
				t.Fatalf("unexpected rewrite: %s", res.Source)
			}
		})
	}
}
