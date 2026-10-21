package infer_test

import "testing"

// A static local holds its initializer or a value a former call assigned;
// any other kind of write leaves it unknown.
func TestStaticLocals(t *testing.T) {
	checkAnywhere(t, `<?php
function f() {
    static $x = false;
    t('static', $x);
    if ($x === false) { $x = 1; }
    t('assigned', $x);
}
function g() {
    static $cache = null;
    if ($cache === null) { $cache = new ArrayObject(); }
    return $cache;
}
function h($k) {
    static $memo = [];
    $memo[$k] = 1;
    t('elem', $memo);
    static $none;
    t('none', $none);
    static $inc = 0;
    $inc++;
    t('inc', $inc);
    static $self = 0;
    $self = $self + 1;
}
function i($k) {
    static $dyn = 1;
    $$k = 2;
    t('dyn', $dyn);

    static $ref = 1;
    $alias = &$ref;
    t('ref', $ref);
}
function j() { t('return', g()); }
function k($c) {
    static $both = 0;
    t('both', $both);
    if ($c) { $both = 'a'; } else { $both = 'b'; }
    static $opaque = 0;
    t('opaque', $opaque);
    $opaque = unknownfn();
    global $shared;
    static $shared = 1;
    t('shared', $shared);
}
`, map[string]string{"both": "int|string", "opaque": "?unknown", "shared": "?unknown", "static": "false|int", "assigned": "int", "elem": "?unknown", "none": "null", "inc": "?unknown", "dyn": "?unknown", "ref": "?unknown", "return": `\ArrayObject`})
}
