<?php
// Before PHP 8.0 strpos() warns and returns false on an empty needle, so only
// a non-empty literal is rewritten.
function prefixes(string $uri, string $base) {
    $r = [];
    $r[] = substr($uri, 0, strlen($base)) === $base;
    $r[] = substr($uri, 0, 0) === '';
    $r[] = substr($uri, 0, strlen("$base")) === "$base";
    $r[] = strpos($uri, '/api') === 0;
    $r[] = strpos($uri, '0x1A') === 0;
    return $r;
}
