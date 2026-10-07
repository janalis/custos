<?php
function words($code, $sep) {
    $r = [];
    // trimming before/after ucfirst-like calls is not equivalent
    $r[] = \trim(lcfirst($code), '#/');
    $r[] = mb_substr(ucwords($code), -2);
    $r[] = trim(ucfirst($code));
    $r[] = rtrim(ucwords($code, '-'), '.');
    // ucfirst/lcfirst/ucwords do not override each other
    $r[] = ucfirst(lcfirst($code));
    $r[] = lcfirst(ucwords($code));
    $r[] = ucwords(ucfirst($code));
    // ucwords with other delimiters than the outer one
    $r[] = ucwords(ucwords($code, '-'));
    $r[] = ucwords(ucwords($code, '-'), '_');
    $r[] = ucwords(<weak_warning descr="The inner 'ucwords(...)' call has no effect here.">ucwords($code, '-')</weak_warning>, '-');
    $r[] = ucwords(<weak_warning descr="The inner 'ucwords(...)' call has no effect here.">ucwords($code)</weak_warning>);
    $r[] = ucfirst(<weak_warning descr="The inner 'ucfirst(...)' call has no effect here.">ucfirst($code)</weak_warning>);
    // a basic case conversion overrides them
    $r[] = mb_strtoupper(<weak_warning descr="The inner 'lcfirst(...)' call has no effect here.">lcfirst($code)</weak_warning>);
    $r[] = <weak_warning descr="Cut first, then change the case: 'strtolower(substr($code, 1))'.">substr(strtolower($code), 1)</weak_warning>;
    return $r;
}
