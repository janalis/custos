<?php
function routes($uri, $base, $o) {
    $r = [];
    $r[] = substr($uri, 1, strlen($base)) == $base;
    $r[] = substr($uri, 00, strlen($base)) == $base;
    $r[] = substr($uri, 0, strpos($uri, '?')) == $base;
    $r[] = trim(substr($uri, 0, strlen($base))) == $base;
    $r[] = (substr($uri, 0, strlen($base))) == $base;
    $r[] = substr($uri, 0, strlen($base)) < $base;
    $r[] = substr($uri, 0) == $base;
    $r[] = mb_convert_case(substr($uri, 0, 3), MB_CASE_LOWER) == $base;
    $r[] = $o->strtolower(substr($uri, 0, 3)) == $base;
    $r[] = $o->substr($uri, 0, 3) == $base;
    $r[] = substr($uri, 0, strlen($base, 1)) == $base;        // strlen() takes one argument
    $r[] = substr($uri, 0, 0x3) == 'abc';                       // not a decimal literal
    $r[] = mb_convert_case(substr($uri, 0, 3)) == 'abc';        // missing mode: never folded
    $x = substr($uri, 0, 3);
    return $r;
}
