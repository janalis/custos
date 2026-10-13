<?php
// The cut length must be the compared value's length, in the same unit.
function prefixes($uri, $base, $other, $n) {
    $r = [];
    $r[] = substr($uri, 0, 3) == $base;
    $r[] = substr($uri, 0, strlen($other)) == $base;
    $r[] = substr($uri, 0, $n) === $base;
    $r[] = substr($uri, 0, 5) === '/api';
    $r[] = substr($uri, 0, mb_strlen($base)) == $base;
    $r[] = mb_substr($uri, 0, strlen($base)) == $base;
    $r[] = mb_substr($uri, 0, 2) == 'éa';
    return $r;
}
