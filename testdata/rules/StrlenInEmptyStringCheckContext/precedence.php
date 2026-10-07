<?php
function pick(string $p, string $q, $raw, ?string $alt, string $tail)
{
    $a = <weak_warning descr="Compare with an empty string instead: '($p ?: $q) !== '''.">strlen($p ?: $q) > 0</weak_warning>;
    $b = <weak_warning descr="Compare with an empty string instead: '(string)($raw ?? $alt) === '''.">!strlen($raw ?? $alt)</weak_warning>;
    $c = <weak_warning descr="Compare with an empty string instead: '(string)($raw + 1) !== '''.">mb_strlen($raw + 1) != 0</weak_warning>;
    $d = <weak_warning descr="Compare with an empty string instead: '$p . $tail !== '''.">strlen($p . $tail) > 0</weak_warning>;
    $e = <weak_warning descr="Compare with an empty string instead: '($p ? $q : $tail) === '''.">strlen($p ? $q : $tail) === 0</weak_warning>;
    $f = <weak_warning descr="Compare with an empty string instead: 'trim((string) $raw) !== '''.">strlen(trim((string) $raw)) >= 1</weak_warning>;
    return [$a, $b, $c, $d, $e, $f];
}
