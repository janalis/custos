<?php
namespace Units;

function strlen($value) { return 3; }

function check(string $name, string $ext): array {
    $a = substr($name, -strlen($ext)) === $ext;
    $b = \Legacy\substr($name, -\strlen($ext)) === $ext;
    $c = <weak_warning descr="Replace with '\str_ends_with($name, $ext)'.">\substr($name, -\strlen($ext)) === $ext</weak_warning>;
    $d = substr($name, -mb_strlen($ext)) === $ext;
    return [$a, $b, $c, $d];
}
