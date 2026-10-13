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
		// Writes are read in order: request text stored after a replacement
		// fills the element again, a replacement after it empties it again.
		{"element removed", "<?php unset($_GET['v']); echo '<p>' . $_GET['v'];", 0},
		{"element removed, table refilled", "<?php unset($_GET['v']); $_GET += $_POST; echo '<p>' . $_GET['v'];", 1},
		{"element removed, table reassigned with request text", "<?php unset($_GET['v']); $_GET = $_GET + $_POST; echo '<p>' . $_GET['v'];", 1},
		{"element removed, table refilled by a builtin", "<?php unset($_GET['v']); parse_str($_SERVER['QUERY_STRING'], $_GET); echo '<p>' . $_GET['v'];", 1},
		{"element removed, element refilled", "<?php unset($_GET['v']); $_GET['v'] = $_POST['x']; echo '<p>' . $_GET['v'];", 1},
		{"element removed, other element refilled", "<?php unset($_GET['v']); $_GET['w'] = $_POST['x']; echo '<p>' . $_GET['v'];", 0},
		{"element removed, unknown element refilled", "<?php unset($_GET['v']); $_GET[$i] = $_POST['x']; echo '<p>' . $_GET['v'];", 1},
		{"element removed, other element refilled through a reference", "<?php unset($_GET['v']); $r = &$_GET['w']; $r = $_POST['x']; echo '<p>' . $_GET['v'];", 0},
		{"element removed, refilled through a reference", "<?php unset($_GET['v']); $r = &$_GET['v']; $r = $_POST['x']; echo '<p>' . $_GET['v'];", 1},
		{"element removed, refilled through a reference to an unknown element", "<?php unset($_GET['v']); $r = &$_GET[$i]; $r = $_POST['x']; echo '<p>' . $_GET['v'];", 1},
		{"element removed, refilled through a reference to a nested element", "<?php unset($_GET['v']); $r = &$_GET['v']['k']; $r = $_POST['x']; echo '<p>' . $_GET['v'];", 1},
		{"element removed, refilled through a reference of a reference", "<?php unset($_GET['v']); $r = &$_GET['v']; $q = &$r; $q = $_POST['x']; echo '<p>' . $_GET['v'];", 1},
		{"other element replaced through a reference", "<?php $r = &$_GET['w']; $r = 'x'; echo '<p>' . $_GET['v'];", 1},
		{"element removed, refilled through a reference to the table", "<?php unset($_GET['v']); $r = &$_GET; $r['v'] = $_POST['x']; echo '<p>' . $_GET['v'];", 1},
		{"element removed, nested element refilled", "<?php $_GET['v'] = 'x'; $_GET['v']['k'] = $_POST['x']; echo '<p>' . $_GET['v']['k'];", 1},
		{"element removed, refilled through the globals table", "<?php function f() { $_GET['v'] = 'x'; $GLOBALS['_GET']['v'] = $_POST['x']; echo '<p>' . $_GET['v']; }", 1},
		{"element removed, globals table entry replaced", "<?php function f() { $_GET['v'] = $_POST['x']; $GLOBALS['_GET']['v'] = 'x'; echo '<p>' . $_GET['v']; }", 0},
		{"element removed, refilled by a destructuring", "<?php unset($_GET['v']); [$_GET['v'], list($x)] = [$_POST['x'], [1]]; echo '<p>' . $_GET['v'];", 1},
		{"element replaced by a destructuring", "<?php [$_GET['v'], , $x] = ['a', 'b', 'c']; echo '<p>' . $_GET['v'];", 0},
		{"element removed, refilled as a loop target", "<?php unset($_GET['v']); foreach ($_POST as $_GET['v']) { } echo '<p>' . $_GET['v'];", 1},
		{"element removed, refilled as a loop key", "<?php unset($_GET['v']); foreach ($_POST as $_GET['v'] => $x) { } echo '<p>' . $_GET['v'];", 1},
		{"element replaced as a loop target", "<?php foreach ([1, 2] as $_GET['v']) { } echo '<p>' . $_GET['v'];", 0},
		{"element removed, refilled later in the loop", "<?php unset($_GET['v']); foreach ($l as $k) { echo '<p>' . $_GET['v']; $_GET['v'] = $_POST['x']; }", 1},
		{"element removed, refilled later in an outer loop", "<?php unset($_GET['v']); while ($c) { foreach ($l as $k) { echo '<p>' . $_GET['v']; } $_GET['v'] = $_POST['x']; }", 1},
		{"element removed, refilled after the loop", "<?php unset($_GET['v']); foreach ($l as $k) { echo '<p>' . $_GET['v']; } $_GET['v'] = $_POST['x'];", 0},
		{"element removed later in the loop", "<?php foreach ($l as $k) { echo '<p>' . $_GET['v']; unset($_GET['v']); }", 1},
		{"element removed, refilled after the output", "<?php unset($_GET['v']); echo '<p>' . $_GET['v']; $_GET['v'] = $_POST['x'];", 0},
		{"element removed, function of the project called", "<?php unset($_GET['v']); fill(); echo '<p>' . $_GET['v'];", 1},
		{"element removed, method called", "<?php unset($_GET['v']); $o->fill(); echo '<p>' . $_GET['v'];", 1},
		{"element removed, static method called", "<?php unset($_GET['v']); K::fill(); echo '<p>' . $_GET['v'];", 1},
		{"element removed, object made", "<?php unset($_GET['v']); $o = new K; echo '<p>' . $_GET['v'];", 1},
		{"element removed, PHP function called", "<?php unset($_GET['v']); $n = strlen('a'); echo '<p>' . $_GET['v'];", 0},
		// A call known to store nothing in a request table changes nothing.
		{"element removed, function that writes no table called", "<?php function note($m) { error_log($m); } unset($_GET['v']); note('x'); echo '<p>' . $_GET['v'];", 0},
		{"element removed, function that writes a table called", "<?php function fill() { $_GET['v'] = $_POST['x']; } unset($_GET['v']); fill(); echo '<p>' . $_GET['v'];", 1},
		{"element removed, method taking a reference called", "<?php class K { private function into(&$a, $b) { $a['v'] = $b; } function f() { unset($_GET['v']); $this->into($_GET, $_POST['x']); echo '<p>' . $_GET['v']; } }", 1},
		{"element removed, private method that writes no table called", "<?php class K { private function note($m) { error_log($m); } function f() { unset($_GET['v']); $this->note('x'); echo '<p>' . $_GET['v']; } }", 0},
		{"element removed, private method that writes a table called", "<?php class K { private function fill() { $_GET = $_POST; } function f() { unset($_GET['v']); $this->fill(); echo '<p>' . $_GET['v']; } }", 1},
		{"element removed, static method that writes no table called", "<?php class K { static function note($m) { error_log($m); } } unset($_GET['v']); K::note('x'); echo '<p>' . $_GET['v'];", 0},
		{"element removed, method of an untyped property called", "<?php class K { private $log; function f() { unset($_GET['v']); $this->log->info('x'); echo '<p>' . $_GET['v']; } }", 1},
		{"element removed, PHP class constructed", "<?php unset($_GET['v']); $d = new DateTime(); echo '<p>' . $_GET['v'];", 0},
		{"element removed, method of a PHP class called", "<?php function f(DateTime $d) { unset($_GET['v']); $d->format('Y'); echo '<p>' . $_GET['v']; }", 0},
		{"element removed, class without constructor constructed", "<?php class K { } unset($_GET['v']); $o = new K(); echo '<p>' . $_GET['v'];", 0},
		{"element removed, class whose constructor writes a table constructed", "<?php class K { function __construct() { $_GET['v'] = $_POST['x']; } } unset($_GET['v']); $o = new K(); echo '<p>' . $_GET['v'];", 1},
		{"function of the project called, element removed", "<?php fill(); unset($_GET['v']); echo '<p>' . $_GET['v'];", 0},
		{"element replaced by what a function of the project returns", "<?php $_GET['v'] = clean($_GET['v'], $o->option()); echo '<p>' . $_GET['v'];", 0},
		{"element replaced by what a closure returns", "<?php $f = function () { return other(); }; $_GET['v'] = (function () { return other(); })->bindTo(null) ? 'a' : 'b'; echo '<p>' . $_GET['v'];", 0},
		{"element removed, function of the project called later in the loop", "<?php unset($_GET['v']); foreach ($l as $k) { echo '<p>' . $_GET['v']; fill(); }", 1},
		{"element removed, table extended with literals", "<?php unset($_GET['v']); $_GET += ['v' => 'x']; echo '<p>' . $_GET['v'];", 0},
		{"table refilled, element removed", "<?php $_GET += $_POST; unset($_GET['v']); echo '<p>' . $_GET['v'];", 0},
		{"element removed, refilled, removed", "<?php unset($_GET['v']); $_GET += $_POST; unset($_GET['v']); echo '<p>' . $_GET['v'];", 0},
		{"element refilled, removed", "<?php $_GET['v'] = $_POST['x']; unset($_GET['v']); echo '<p>' . $_GET['v'];", 0},
		{"table emptied", "<?php $_GET = []; echo '<p>' . $_GET['v'];", 0},
		{"table emptied, element refilled", "<?php $_GET = []; $_GET['v'] = $_POST['x']; echo '<p>' . $_GET['v'];", 1},
		{"table emptied, other element refilled", "<?php $_GET = []; $_GET['w'] = $_POST['x']; echo '<p>' . $_GET['v'];", 0},
		{"table emptied, one of two elements refilled", "<?php $_GET = []; $_GET['v'] = $_POST['x']; echo '<p>' . $_GET['v'] . $_GET['w'];", 1},
		{"table emptied, element refilled, table emptied", "<?php $_GET = []; $_GET['v'] = $_POST['x']; $_GET = []; echo '<p>' . $_GET['v'];", 0},
		{"table emptied, element refilled, read whole", "<?php $_GET = []; $_GET['v'] = $_POST['x']; echo '<p>' . implode(',', $_GET);", 0},
		{"table emptied, table refilled", "<?php $_GET = []; $_GET = $_POST; echo '<p>' . $_GET['v'];", 1},
		{"other table emptied, table refilled", "<?php $_POST = []; $_GET += $_COOKIE; echo '<p>' . $_GET['v'] . $_POST['w'];", 1},
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
