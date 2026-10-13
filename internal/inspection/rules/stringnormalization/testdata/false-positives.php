<?php
function tidy($code, $mask, $o) {
    $r = [];
    $r[] = trim(strtoupper($code), 'xyz');
    $r[] = trim(strtoupper($code), " \n");
    $r[] = trim(strtoupper($code), $mask);
    $r[] = $o->rtrim(strtoupper($code));
    $r[] = rtrim($o->strtoupper($code));
    $r[] = strtoupper(rtrim($code));
    $r[] = trim((strtoupper($code)));
    $r[] = substr($code, strtolower($code));
    $r[] = ucwords(strtolower($code));
    $r[] = lcfirst(mb_strtolower($code));
    $r[] = strtolower(mb_strtoupper($code));
    $r[] = strtoupper(ucwords($code, '-'));
    $r[] = trim(strtolower());
    return $r;
}
