<?php
function tidy($code) {
    $r = [];
    $r[] = strtoupper(ltrim($code));
    $r[] = mb_strtolower(substr($code, 0, 3));
    $r[] = mb_convert_case(rtrim($code, '-'), MB_CASE_TITLE);
    $r[] = mb_convert_case(trim($code), $code . 'x');

    $r[] = mb_strtoupper($code);
    $r[] = lcfirst($code);
    $r[] = strtoupper($code);
    $r[] = mb_strtolower($code);
    $r[] = strtolower(trim(strtolower($code)));
    return $r;
}
