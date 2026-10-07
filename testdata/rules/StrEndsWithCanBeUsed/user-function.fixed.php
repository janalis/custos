<?php
namespace Units;

function strlen($value) { return 3; }

function check(string $name, string $ext): array {
    $a = substr($name, -strlen($ext)) === $ext;
    $b = \Legacy\substr($name, -\strlen($ext)) === $ext;
    $c = \str_ends_with($name, $ext);
    $d = str_ends_with($name, $ext);
    return [$a, $b, $c, $d];
}
