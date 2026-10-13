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
    $r[] = ucwords($code, '-');
    $r[] = ucwords($code);
    $r[] = ucfirst($code);
    // a basic case conversion overrides them
    $r[] = mb_strtoupper($code);
    $r[] = strtolower(substr($code, 1));
    return $r;
}
