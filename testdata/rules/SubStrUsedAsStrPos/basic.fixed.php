<?php
function routes($uri, $base, $enc) {
    $r = [];
    $r[] = substr($uri, 0, strlen($base)) == $base;
    $r[] = strpos($uri, '/api') !== 0;
    $r[] = \mb_strpos($uri, $base) === 0;
    $r[] = mb_strpos($uri, $base, 0, $enc) !== 0;
    $r[] = mb_strpos($uri, 'abc', 0, 'UTF-8') === 0;
    $r[] = stripos($uri, 'get:') === 0;
    $r[] = mb_stripos($uri, 'POST') === 0;
    $r[] = substr($uri, 0, strlen($base)) <> $base;
    return $r;
}
