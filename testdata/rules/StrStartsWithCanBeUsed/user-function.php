<?php
namespace Paths;

function mb_strpos($haystack, $needle) { return 0; }

function check(string $path): array {
    $a = mb_strpos($path, '/') === 0;
    $b = \Legacy\strpos($path, '/') === 0;
    $c = <weak_warning descr="Replace with 'str_starts_with($path, '/')'.">strpos($path, '/') === 0</weak_warning>;
    $d = \mb_strpos($path, '.') !== 0;
    return [$a, $b, $c, $d];
}
