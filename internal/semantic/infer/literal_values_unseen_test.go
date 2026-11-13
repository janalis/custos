package infer_test

import "testing"

// A local that may be written where its reaching definitions do not show
// it keeps its atoms but not the values of those definitions.
func TestLiteralValuesDroppedOnUnseenWrites(t *testing.T) {
	checkShape(t, `<?php
function other() { global $g; $g = 5; }
function byRef(&$x) { $x = 9; }
function reference() { $c = 0; $p = &$c; $p = 5; t('reference', $c); }
function arrayReference() { $c = 0; $r = [&$c]; $r[0] = 3; t('arrayReference', $c); }
function elementReference() { $c = 0; $m['x'] = &$c; $m['x'] = 4; t('elementReference', $c); }
function propertyReference(stdClass $obj) { $c = 0; $obj->r = &$c; $obj->r = 4; t('propertyReference', $c); }
function globalAfterCall() { global $g; $g = 0; other(); t('global', $g); }
function staticAfterCall() { static $n = 0; $n = 0; other(); t('static', $n); }
function closureByRef() { $c = 0; $fn = static function (&$x) { $x = 9; }; $fn($c); t('closure', $c); }
function byRefParam() { $c = 0; byRef($c); t('byRefParam', $c); }
function stringReference() { $s = ''; $p = &$s; $p .= 'x'; t('stringReference', $s); }
function plain() { $c = 0; other(); t('plain', $c); }
`, map[string]string{
		"reference":         "int",
		"arrayReference":    "int",
		"elementReference":  "int",
		"propertyReference": "int",
		"global":            "int",
		"static":            "int",
		"closure":           "int",
		"stringReference":   "string",
		"plain":             "int<0, 0>",
	})
}

// A callee named at run time (a dynamic method or static method name, a
// class held by a variable) may take its arguments by reference, like a
// callable called through a variable.
func TestLiteralValuesDroppedOnDynamicCallees(t *testing.T) {
	checkShape(t, `<?php
function method($obj, string $m) { $n = 0; $obj->$m($n); t('method', $n); }
function staticMethod(string $m) { $n = 0; Foo::$m($n); t('staticMethod', $n); }
function dynamicClass(string $cls) { $n = 0; $cls::run($n); t('dynamicClass', $n); }
function construct(string $cls) { $n = 0; new $cls($n); t('construct', $n); }
function callable($fn) { $n = 0; $fn($n); t('callable', $n); }
`, map[string]string{
		"method":       "int",
		"staticMethod": "int",
		"dynamicClass": "int",
		"construct":    "int",
		"callable":     "int",
	})
}

// A local read after an include or eval, which may assign it, keeps no
// value.
func TestLiteralValuesDroppedAfterInclude(t *testing.T) {
	checkShape(t, `<?php
$c = 0;
include 'x.php';
t('include', $c);
function f() { $c = 0; eval('$c = 1;'); t('eval', $c); }
function g() { $c = 0; require 'x.php'; t('require', $c); }
`, map[string]string{"include": "?unknown", "eval": "?unknown", "require": "?unknown"})
}
