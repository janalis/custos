<?php
function pick(string $p, string $q, $raw, ?string $alt, string $tail)
{
    $a = ($p ?: $q) !== '';
    $b = (string)($raw ?? $alt) === '';
    $c = mb_strlen($raw + 1) != 0;
    $d = $p . $tail !== '';
    $e = ($p ? $q : $tail) === '';
    $f = trim((string) $raw) !== '';
    return [$a, $b, $c, $d, $e, $f];
}
