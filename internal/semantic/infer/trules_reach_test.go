package infer

import (
	"testing"

	phpversion "custos/internal/php/version"
)

// An unconditional reassignment hides earlier assignments from the
// T-rules (outside SpecOnly): `$s = undeclared($s)` is unknown, not int.
func TestTRulesReachingDefinitions(t *testing.T) {
	src := `<?php
function f(int $p, bool $c) {
    $s = 0; $s = undeclared($s); t('killed', $s);
    $p = undeclared(); t('param', $p);
    $u = 0; if ($c) { $u = 'a'; } t('branch', $u);
    $w = 1; $w = 'x'; t('last', $w);
}
`
	checkLabels(t, trulesLabels(t, src, phpversion.PHP84, nil), map[string]string{"killed": "", "param": "", "branch": "int|string", "last": "string"})
	checkLabels(t, trulesLabels(t, src, phpversion.PHP84, func(r *TRules) { r.SpecOnly = true }), map[string]string{"killed": "int", "last": "int|string"})
}

// Web SAPIs store the ports as strings.
func TestTRulesServerPorts(t *testing.T) {
	src := `<?php
t('port', $_SERVER['SERVER_PORT']); t('remote', $_SERVER['REMOTE_PORT']); t('time', $_SERVER['REQUEST_TIME']);
`
	checkLabels(t, trulesLabels(t, src, phpversion.PHP84, nil), map[string]string{"port": "string", "remote": "string", "time": "int"})
}

// Inline annotations: one on an assignment keeps the assigned value's
// T-rules type; a standalone one states the type.
func TestTRulesReachingAnnotations(t *testing.T) {
	src := `<?php
function f() {
    /** @var mixed $a */
    $a = 'x'; t('annotated', $a);
    static $z;
    if ($z === null) { $z = 1; }
    /** @var string $z */
    t('standalone', $z);
}
`
	checkLabels(t, trulesLabels(t, src, phpversion.PHP84, nil), map[string]string{"annotated": "string", "standalone": "string"})
}
