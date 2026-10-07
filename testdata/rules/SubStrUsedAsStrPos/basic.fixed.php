<?php
function routes($uri, $base, $enc) {
    $r = [];
    $r[] = strpos($uri, $base) === 0;
    $r[] = strpos($uri, '/api') !== 0;
    $r[] = \mb_strpos($uri, $base) === 0;
    $r[] = mb_strpos($uri, $base, 0, $enc) !== 0;
    $r[] = mb_strpos($uri, 'abc', 0, 'UTF-8') === 0;
    $r[] = stripos($uri, 'get:') === 0;
    $r[] = mb_stripos($uri, 'POST') === 0;
    $r[] = strpos($uri, $base) !== 0;
    return $r;
}
