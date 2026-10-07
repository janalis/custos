<?php
namespace Paths;

function mb_strpos($haystack, $needle) { return 0; }

function check(string $path): array {
    $a = mb_strpos($path, '/') === 0;
    $b = \Legacy\strpos($path, '/') === 0;
    $c = str_starts_with($path, '/');
    $d = !\str_starts_with($path, '.');
    return [$a, $b, $c, $d];
}
