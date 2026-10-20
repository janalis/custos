package untrustedheadervalue

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
	const read = "$v = $_GET['v']; "
	const send = " header('X-L: ' . $v);"
	for _, tc := range []struct {
		name, source string
		findings     int
	}{
		// Line-break rejections that leave before the header.
		{"truthy match", read + `if (preg_match('/[\r\n]/', $v)) { exit; }` + send, 0},
		{"match equals one", read + `if (preg_match('/[\r\n]/', $v) === 1) { exit; }` + send, 0},
		{"match loosely one", read + `if (preg_match('/[\r\n]/', $v) == 1) die();` + send, 0},
		{"match not zero", read + `if (preg_match('/[\r\n]/', $v) !== 0) { return; }` + send, 0},
		{"match loosely not zero", read + `if (preg_match('/[\r\n]/', $v) != 0) { return; }` + send, 0},
		{"match above zero", read + `if (preg_match('/[\r\n]/', $v) > 0) { exit; }` + send, 0},
		{"match not false", read + `if (preg_match('/[\r\n]/', $v) != false) { exit; }` + send, 0},
		{"yoda", read + `if (1 === preg_match('/[\r\n]/', $v)) { exit; }` + send, 0},
		{"double quoted pattern", read + `if (preg_match("/[\r\n]/", $v)) { exit; }` + send, 0},
		{"other delimiter", read + `if (preg_match('~[\n\r]~', $v)) { exit; }` + send, 0},
		{"named arguments", read + `if (preg_match(subject: $v, pattern: '/[\r\n]/')) { exit; }` + send, 0},
		{"saved test", read + `$bad = preg_match('/[\r\n]/', $v); if ($bad) { exit; }` + send, 0},
		{"exit with else", read + `if (preg_match('/[\r\n]/', $v)) { exit; } else { $x = 1; }` + send, 0},
		{"elseif after failed test", read + `if (preg_match('/[\r\n]/', $v)) { exit; } elseif ($x) {` + send + ` }`, 0},
		{"short circuit", read + `preg_match('/[\r\n]/', $v) or header('X-L: ' . $v);`, 0},
		{"match arm", read + `echo match (true) { strpbrk($v, "\r\n") === false => header('X-L: ' . $v), default => 0 };`, 0},
		{"jump label after guard", read + `if (preg_match('/[\r\n]/', $v)) { exit; } back:` + send + ` goto back;`, 1},
		{"strpbrk strict", read + `if (strpbrk($v, "\r\n") !== false) { exit; }` + send, 0},
		{"strpbrk truthy", read + `if (strpbrk($v, "\n\r\0")) { exit; }` + send, 0},
		{"disjunction", read + `if ($v === '' || preg_match('/[\r\n]/', $v)) { exit; }` + send, 0},
		{"disjunction first", read + `if (preg_match('/[\r\n]/', $v) or $v === '') { exit; }` + send, 0},
		{"double negation", read + `if (!!preg_match('/[\r\n]/', $v)) { exit; }` + send, 0},
		{"guard on copy", read + `$w = $v; if (preg_match('/[\r\n]/', $w)) { exit; }` + send, 0},
		{"switch case", read + `switch ($m) { case 1: if (preg_match('/[\r\n]/', $v)) { exit; }` + send + ` }`, 0},
		{"loop", `foreach ($xs as $x) { ` + read + `if (preg_match('/[\r\n]/', $v)) { continue; }` + send + ` }`, 0},
		{"function scope", `function f() { ` + read + `if (preg_match('/[\r\n]/', $v)) { return; }` + send + ` }`, 0},
		{"nested use", read + `if (preg_match('/[\r\n]/', $v)) { exit; } if ($m) {` + send + ` }`, 0},
		// Branches entered only without a line break.
		{"negated branch", read + `if (!preg_match('/[\r\n]/', $v)) {` + send + ` }`, 0},
		{"not one branch", read + `if (preg_match('/[\r\n]/', $v) !== 1) {` + send + ` }`, 0},
		{"loosely zero branch", read + `if (preg_match('/[\r\n]/', $v) == 0) {` + send + ` }`, 0},
		{"strpbrk false branch", read + `if (strpbrk($v, "\r\n") === \false) {` + send + ` }`, 0},
		{"conjunction branch", read + `if ($m && !preg_match('/[\r\n]/', $v)) {` + send + ` }`, 0},
		{"conjunction first", read + `if (!preg_match('/[\r\n]/', $v) and $m) {` + send + ` }`, 0},
		{"else branch", read + `if (preg_match('/[\r\n]/', $v)) { echo 'bad'; } else {` + send + ` }`, 0},
		// Guards that prove nothing.
		{"zero leaves", read + `if (preg_match('/[\r\n]/', $v) === 0) { exit; }` + send, 1},
		{"below one leaves", read + `if (preg_match('/[\r\n]/', $v) < 1) { exit; }` + send, 1},
		{"above one", read + `if (preg_match('/[\r\n]/', $v) > 1) { exit; }` + send, 1},
		{"above variable", read + `if (preg_match('/[\r\n]/', $v) > $m) { exit; }` + send, 1},
		{"strpbrk above zero", read + `if (strpbrk($v, "\r\n") > 0) { exit; }` + send, 1},
		{"identical true", read + `if (preg_match('/[\r\n]/', $v) === true) { exit; }` + send, 1},
		{"identical two", read + `if (preg_match('/[\r\n]/', $v) === 2) { exit; }` + send, 1},
		{"identical variable", read + `if (preg_match('/[\r\n]/', $v) === $m) { exit; }` + send, 1},
		{"strpbrk equals one", read + `if (strpbrk($v, "\r\n") == 1) { exit; }` + send, 1},
		{"bitwise not", read + `if (~preg_match('/[\r\n]/', $v)) { exit; }` + send, 1},
		{"variable condition", read + `if ($m) { exit; }` + send, 1},
		{"conjunction leaves", read + `if ($m && preg_match('/[\r\n]/', $v)) { exit; }` + send, 1},
		{"disjunction branch", read + `if ($m || !preg_match('/[\r\n]/', $v)) {` + send + ` }`, 1},
		{"positive branch", read + `if (preg_match('/[\r\n]/', $v)) {` + send + ` }`, 1},
		{"else of negation", read + `if (!preg_match('/[\r\n]/', $v)) { echo 'ok'; } else {` + send + ` }`, 1},
		{"compared values", read + `if ($m === false) { exit; }` + send, 1},
		{"negated class", read + `if (preg_match('/[^\r\n]/', $v)) { exit; }` + send, 1},
		{"modifier", read + `if (preg_match('/[\r\n]/u', $v)) { exit; }` + send, 1},
		{"letter delimiter", read + `if (preg_match('![\r\n]!', $v)) { exit; }` + send, 1},
		{"no class", read + `if (preg_match('/\r|\n/', $v)) { exit; }` + send, 1},
		{"short pattern", read + `if (preg_match('//', $v)) { exit; }` + send, 1},
		{"escaped backslash", read + `if (preg_match('/[\\\\r\n]/', $v)) { exit; }` + send, 1},
		{"only carriage return", read + `if (preg_match('/[\r]/', $v)) { exit; }` + send, 1},
		{"strpbrk one byte", read + `if (strpbrk($v, "\n") !== false) { exit; }` + send, 1},
		{"dynamic pattern", read + `if (preg_match($p, $v)) { exit; }` + send, 1},
		{"offset argument", read + `if (preg_match('/[\r\n]/', $v, $m, 0, 3)) { exit; }` + send, 1},
		{"other builtin", read + `if (str_starts_with($v, "\r\n")) { exit; }` + send, 1},
		{"other variable", read + `if (preg_match('/[\r\n]/', $w)) { exit; }` + send, 1},
		{"trimmed subject", read + `if (preg_match('/[\r\n]/', trim($v))) { exit; }` + send, 1},
		{"guard falls through", read + `if (preg_match('/[\r\n]/', $v)) { log_it(); }` + send, 1},
		{"reassigned", read + `if (preg_match('/[\r\n]/', $v)) { exit; } $v = $_GET['w'];` + send, 1},
		{"conditional reassignment", read + `if (preg_match('/[\r\n]/', $v)) { exit; } if ($m) { $v = $_GET['w']; }` + send, 1},
		{"guard after use", read + `header('X-L: ' . $v); if (preg_match('/[\r\n]/', $v)) { exit; }`, 1},
		{"guard in sibling branch", read + `if ($m) { if (preg_match('/[\r\n]/', $v)) { exit; } }` + send, 1},
		{"two origins", `$v = $m ? $_GET['a'] : $_GET['b']; if (preg_match('/[\r\n]/', $v)) { exit; }` + send, 1},
		// Values that cannot hold request bytes.
		{"length", `header('Content-Length: ' . strlen($_POST['body']));`, 0},
		{"count", `header('X-Count: ' . count($_GET['ids']));`, 0},
		{"type test", `header('X-Numeric: ' . is_numeric($_GET['v']));`, 0},
		{"character class test", `header('X-Digits: ' . ctype_digit($_GET['v']));`, 0},
		{"assigned length", `$n = strlen($_POST['body']); header("Content-Length: $n");`, 0},
		{"arithmetic", `header('X-A: ' . ($_GET['a'] - 1) . ($_GET['a'] * 2) . ($_GET['a'] / 2) . ($_GET['a'] % 2) . ($_GET['a'] ** 2));`, 0},
		{"sign", `header('X-A: ' . -$_GET['a'] . +$_GET['a']);`, 0},
		{"addition with literal", `header('X-Page: ' . ($_GET['page'] + 1));`, 0},
		{"addition with cast", `header('X-Sum: ' . ((string) $_GET['a'] + $_GET['b']));`, 0},
		{"addition with length", `header('X-Sum: ' . ($_GET['a'] + strlen($_GET['b'])));`, 0},
		{"match literals", `$d = match ($_GET['sort']) { 'desc' => 'DESC', default => 'ASC' }; header('X-Sort: ' . $d);`, 0},
		{"match throwing arm", `header('X-Sort: ' . match ($_GET['sort']) { 'desc' => 'DESC', default => throw new Exception('sort') });`, 0},
		{"ternary lengths", `header('X-N: ' . ($m ? strlen($_GET['a']) : count($_GET['b'])));`, 0},
		{"short ternary", `header('X-N: ' . (strlen($_GET['a']) ?: 'none'));`, 0},
		// Values that keep request bytes.
		{"array union", `header('X-Ids: ' . implode(',', $_GET['a'] + $_GET['b']));`, 1},
		{"addition of complement", `header('X-Ids: ' . implode(',', ~$_GET['a'] + $_GET['b']));`, 1},
		{"match request arm", `header('X-V: ' . match ($_GET['m']) { 'raw' => $_GET['v'], default => 'none' });`, 1},
		{"ternary request arm", `header('X-V: ' . ($m ? strlen($_GET['a']) : $_GET['b']));`, 1},
		{"short ternary request", `header('X-V: ' . ($_GET['a'] ?: strlen($_GET['b'])));`, 1},
		{"interpolated request", `$n = strlen($_GET['a']); header("X-V: $n {$_GET['b']}");`, 1},
		{"coalesce", `header('X-V: ' . ($_GET['a'] ?? strlen($_GET['b'])));`, 1},
		{"complement", `header('X-V: ' . ~$_GET['a'] . strlen($_GET['b']));`, 1},
		{"string function", `header('X-V: ' . trim($_GET['a']) . strlen($_GET['b']));`, 1},
		{"match numeric sum arm", `header('X-V: ' . match ($_GET['k']) { 'a' => $_GET['n'] + 1, default => 0 });`, 0},
		{"depth bound", `header(($_GET['a'] + 1)` + strings.Repeat(" . 'x'", 40) + `);`, 1},
		// Calls the local proofs do not apply to.
		{"spread", `header(...$_GET['args']);`, 0},
		{"wrapper", `function send($v) { header('X-L: ' . $v); } send($_GET['v']);`, 1},
		{"closure scope", read + `if (preg_match('/[\r\n]/', $v)) { exit; } $f = function () { $v = $_GET['v'];` + send + ` };`, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("review.php", []byte("<?php "+tc.source), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
