package flow

import (
	"fmt"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

// What was established of a static variable is gone past a call of PHP that
// runs code of the project by what it is given: a class it loads by name, a
// handler, a class it constructs, a callback an option names. It holds
// across the same functions given nothing of the kind.
func TestStaticGuardAcrossCodePHPRuns(t *testing.T) {
	const src = `function probe($x) {}
		function mine($x) { return $x; }
		final class K {
			function t($a, $o, string $s, array $l, PDOStatement $st, mysqli_result $res, SoapServer $soap, LimitIterator $it, ArrayIterator $plain, $k) { static $u; $u = parse_url($a); if ($u === false) { return; } %s probe($u); }
		}`
	for _, tc := range []struct {
		body string
		kept bool
	}{
		// A function named as a callback runs nothing only when it is one
		// of those that compute on text and numbers.
		{`$t = call_user_func('strlen', $s);`, true},
		{`$t = call_user_func('\\strlen', $s);`, true},
		{`$t = call_user_func('STRTOUPPER', $s);`, true},
		{`$t = call_user_func('method_exists', $s, 'm');`, false},
		{`$t = call_user_func('defined', $s);`, false},
		{`$t = call_user_func('ob_get_clean');`, false},
		{`$t = call_user_func('session_encode');`, false},
		{`$t = call_user_func('extract', ['u' => false]);`, false},
		{`$t = call_user_func('mine', $s);`, false},
		{`$t = call_user_func('missing_function', $s);`, false},
		{`$t = array_map('get_parent_class', explode(',', $s));`, false},
		{`$t = array_map('trim', explode(',', $s)); $t = [];`, true},
		// A class named by text is loaded; the class of the code that runs
		// is loaded already. An object handed to PHP may be converted.
		{`$t = method_exists($s, 'm');`, false},
		{`$t = method_exists('Missing', 'm');`, false},
		{`$t = method_exists($this, 'm');`, false},
		{`$t = method_exists(($this), 'm');`, false},
		{`$t = method_exists(self::class, 'm');`, true},
		{`$t = method_exists(static::class, 'm');`, true},
		{`$t = method_exists(Other::class, 'm');`, false},
		{`$t = method_exists(self::OTHER, 'm');`, false},
		{`$t = method_exists($$s, 'm');`, false},
		{`$t = property_exists(new ArrayObject, 'p');`, false},
		{`$t = is_subclass_of(clone $this, 'B');`, false},
		{`$t = is_subclass_of($s, 'B');`, false},
		{`$t = is_subclass_of(object_or_class: $s, class: 'B');`, false},
		{`$t = get_parent_class();`, true},
		{`$t = get_parent_class($s);`, false},
		{`$t = class_implements($s);`, false},
		{`$t = class_parents($s);`, false},
		{`$t = class_uses($s);`, false},
		{`$t = get_class_methods($s);`, false},
		{`$t = get_class_vars($s);`, false},
		{`$t = is_a($s, 'B', true);`, false},
		{`$t = is_a($s, 'B', $k);`, false},
		{`$t = is_a($s, 'B');`, true},
		{`$t = is_a($s, 'B', false);`, true},
		{`$t = is_a($this, 'B', true);`, false},
		{`$t = is_a(...$l);`, false},
		// A name written without "::" names no class.
		{`$t = is_callable('strlen');`, true},
		{`$t = is_callable(function () { return 1; });`, true},
		{`$t = is_callable(fn () => 1);`, true},
		{`$t = is_callable('Missing::m');`, false},
		{`$t = is_callable($s);`, false},
		{`$t = is_callable();`, true},
		{`$t = defined('SOME_NAME');`, true},
		{`$t = defined('Missing::NAME');`, false},
		{`$t = defined($s);`, false},
		{`$t = constant('PHP_EOL');`, true},
		{`$t = constant($s);`, false},
		{`$t = constant(...$l);`, false},
		// Handlers of the project.
		{`spl_autoload($s);`, false},
		{`ob_end_flush();`, false},
		{`$t = ob_get_clean();`, false},
		{`ob_flush();`, false},
		{`ob_clean();`, false},
		{`ob_end_clean();`, false},
		{`$t = ob_get_flush();`, false},
		{`$t = ob_get_level();`, true},
		{`session_abort();`, false},
		{`$t = session_create_id();`, false},
		{`$t = session_encode();`, false},
		{`$t = session_name();`, true},
		{`$t = curl_exec($o);`, false},
		{`$t = xml_parse($o, $s);`, false},
		{`$t = preg_replace_callback_array(['/a/' => 'mine'], $s);`, false},
		// A class to construct, a filter that calls back.
		{`$t = simplexml_load_string($s, 'Mine'); $t = null;`, false},
		{`$t = simplexml_load_file($s, $k); $t = null;`, false},
		{`$t = simplexml_load_string(data: $s); $t = null;`, false},
		{`$t = mysqli_fetch_object($res, 'Mine'); $t = null;`, false},
		{`$t = mysqli_fetch_object(...$l); $t = null;`, false},
		{`$t = pg_fetch_object($o, 0, 'Mine'); $t = null;`, false},
		{`$t = pg_fetch_object(...$l); $t = null;`, false},
		{`$t = filter_var($s, FILTER_VALIDATE_INT);`, true},
		{`$t = filter_var($s, FILTER_VALIDATE_INT, ['flags' => FILTER_FLAG_ALLOW_HEX | (FILTER_NULL_ON_FAILURE), 'options' => ['min_range' => 1, 'max_range' => K::class]]);`, true},
		{`$t = filter_var($s);`, true},
		{`$t = filter_var($s, FILTER_CALLBACK, ['options' => 'mine']);`, false},
		{`$t = filter_var($s, \FILTER_CALLBACK, ['options' => $k]);`, false},
		{`$t = filter_var($s, $k);`, false},
		{`$t = filter_var(...$l);`, false},
		{`$t = filter_input(INPUT_GET, 'a', FILTER_VALIDATE_INT);`, true},
		{`$t = filter_input(INPUT_GET, 'a', FILTER_CALLBACK, ['options' => 'mine']);`, false},
		{`$t = filter_input(INPUT_GET, 'a', $k);`, false},
		{`$t = filter_input(...$l);`, false},
		{`$t = filter_var_array(['a' => $s], $k);`, false},
		{`$t = filter_input_array(INPUT_GET, $k);`, false},
		// Methods of PHP that construct a class, call a function or move
		// an iterator that wraps another.
		{`$t = $st->fetchObject('Mine'); $t = null;`, false},
		{`$t = $st->fetchObject(...$l); $t = null;`, false},
		{`$t = $st->fetchAll(PDO::FETCH_CLASS, 'Mine'); $t = null;`, false},
		{`$t = $st->fetchAll(...$l); $t = null;`, false},
		{`$t = $st->rowCount();`, true},
		{`$t = $res->fetch_object('Mine'); $t = null;`, false},
		{`$soap->handle();`, false},
		{`$it->rewind();`, false},
		{`$plain->rewind();`, true},
		{`$t = new ReflectionClass($s); $t = null;`, false},
		{`$t = new ReflectionMethod($s, 'm'); $t = null;`, false},
		{`$t = new ReflectionClass(...$l); $t = null;`, false},
		{`$t = new ArrayObject([], 0, $s); $t = null;`, false},
		{`$t = new ArrayObject(...$l); $t = null;`, false},
	} {
		t.Run(tc.body, func(t *testing.T) {
			e, f := environmentAt(t, fmt.Sprintf(src, tc.body), phpversion.PHP84)
			calls := callsNamed(f, "probe")
			if got := e.Excludes(argument(calls[len(calls)-1]), "false"); got != tc.kept {
				t.Fatalf("non-false=%v", got)
			}
		})
	}
}

// The closed list names functions only: a method or a text that is no
// function's name is not one of them, and a call written without its
// argument list gives nothing.
func TestCodePHPRunsEdges(t *testing.T) {
	if pure("extract") || !pure("strlen") || runner("strlen") || !runner("spl_autoload") {
		t.Fatal("pure and runner disagree with their lists")
	}
	if loads("method_exists", nil) || loads("strlen", nil) || invokes("datetime::format", nil) || builds("datetime", nil) {
		t.Fatal("a call given nothing loads nothing")
	}
	if !invokes(`\soapserver::handle`, nil) || !invokes(`pdostatement::fetchall`, &syntax.ArgList{Args: []syntax.Expr{&syntax.Arg{Name: &syntax.Identifier{Value: "mode"}}}}) {
		t.Fatal("a method that serves a request, or one given arguments by name")
	}
	e, _ := environment(t, `function f() {}`, nil)
	if e.keyed(&syntax.FuncCall{}, "json_decode", nil) {
		t.Fatal("JSON decoded without arguments")
	}
}

// A PHP method may call a method a class of the project replaces: it is not
// known to spare the request tables once a class of the project extends its
// class, unless it is final. A construction makes an object of that very
// class.
func TestSparesMethodsOfExtendedPHPClasses(t *testing.T) {
	const declared = `
		function refill($v) { $_GET['a'] = $v; return $v; }
		function refill_then_reads($v) { $_GET['a'] = file_get_contents('php://input'); $w = 1; strlen($v); return $w; }
		function computed($v) { ${'_GET'}['a'] = $v; }
		function by_default($o = new Loud()) { return 1; }
		function quiet_default($o = null, $p = 2) { return 1; }
		function reads($v) { return strlen($v); }
		function flushes() { ob_end_flush(); }
		function checks($s) { return method_exists($s, 'm'); }
		function fetches(PDOStatement $st) { return $st->fetchObject('Loud'); }
		function reflects($s) { return new ReflectionClass($s); }
		class Loud { function __construct() { $_GET['a'] = $_POST['a']; } }
		final class Promoted { function __construct(public Loud $l = new Loud()) { } }
		class Heap extends SplMinHeap { protected function compare($a, $b): int { $_GET['a'] = $_POST['a']; return 0; } }
		class Failure extends RuntimeException { }
	`
	for _, tc := range []struct {
		name, call string
		spared     bool
	}{
		{"method of a PHP class the project extends", `$h->insert(1);`, false},
		{"method of a PHP class nothing extends", `$d->format('Y');`, true},
		{"final method of a PHP class the project extends", `$x->getMessage();`, true},
		{"PHP class the project extends, constructed", `new SplMinHeap();`, true},
		{"reflection constructed", `new ReflectionClass($a);`, false},
		{"array object given a class of iterators", `new ArrayObject([], 0, $a);`, false},
		{"array object", `new ArrayObject([]);`, true},
		{"method that constructs a class", `$st->fetchObject('Loud');`, false},
		{"method that fetches rows", `$st->fetch();`, true},
		{"iterator that wraps another", `$it->rewind();`, false},
		{"PHP function that loads a class by name", `method_exists($a, 'm');`, false},
		{"PHP function given a variable", `method_exists($d, 'm');`, false},
		{"PHP function given the object itself", `method_exists($this, 'm');`, true},
		{"PHP function given the object itself in parentheses", `property_exists(($this), 'p');`, true},
		{"PHP function given a new object", `get_class_methods(new ArrayObject([]));`, true},
		{"PHP function given a copy", `is_subclass_of(clone $d, 'B');`, true},
		{"PHP function given the class of the code", `get_class_vars(self::class);`, true},
		{"PHP function asked whether text names a class", `is_a($this, 'B', true);`, true},
		{"PHP function that flushes a buffer", `ob_end_flush();`, false},
		{"function that stores what it is given", `refill($a);`, false},
		{"function that stores, then reads a variable and calls PHP", `refill_then_reads($a);`, false},
		{"function that names a table by text", `computed($a);`, false},
		{"function whose default value constructs", `by_default();`, false},
		{"function whose default values are constants", `quiet_default();`, true},
		{"constructor whose promoted default constructs", `new Promoted();`, false},
		{"function that only reads", `reads($a);`, true},
		{"function that flushes a buffer", `flushes();`, false},
		{"function that loads a class by name", `checks($a);`, false},
		{"function that fetches an object", `fetches($st);`, false},
		{"function that reflects a class", `reflects($a);`, false},
		{"function given an array that shares a variable", `reads([&$a]);`, false},
		{"function given an array that shares one deeper", `reads(['k' => [1, &$a]]);`, false},
		{"function given an array", `reads([$a]);`, true},
		{"function made into a closure", `refill(...);`, true},
		{"method made into a closure", `$h->insert(...);`, true},
		{"static method made into a closure", `Closure::fromCallable(...);`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := declared + `function t($a, Heap $h, DateTime $d, Failure $x, PDOStatement $st, LimitIterator $it) { ` + tc.call + ` }`
			f := syntax.Parse("test.php", []byte("<?php "+src), syntax.Options{Version: phpversion.Max})
			if len(f.Errors) > 0 {
				t.Fatalf("parse: %v", f.Errors)
			}
			ix := index.New(stubs.Index())
			ix.Add(index.Extract(f))
			types := infer.NewEnv(f, names.New(f), ix, phpversion.Max)
			_, own := Summarize(f, types, NewSnapshot())
			e := New(f, types, NewSnapshot(own))
			call := f.Stmts[len(f.Stmts)-1].(*syntax.Function).Body.Stmts[0].(*syntax.ExprStmt).Expr
			for range 2 {
				// The second answer is the one kept.
				if got := e.Spares(call); got != tc.spared {
					t.Fatalf("spares=%v", got)
				}
			}
		})
	}
}
