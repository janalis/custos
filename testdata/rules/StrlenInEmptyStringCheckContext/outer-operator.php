<?php
function outer($s, string $t, $u, bool $on)
{
    $r = [];
    $r[] = !<weak_warning descr="Compare with an empty string instead: '(string)$s === '''.">!strlen($s)</weak_warning>;
    $r[] = (int) <weak_warning descr="Compare with an empty string instead: '$t === '''.">!strlen($t)</weak_warning>;
    $r[] = <weak_warning descr="Compare with an empty string instead: '(string)$s === '''.">!strlen($s)</weak_warning> . 'x';
    $r[] = <weak_warning descr="Compare with an empty string instead: '(string)$s === '''.">!strlen($s)</weak_warning> < 1;
    $r[] = <weak_warning descr="Compare with an empty string instead: '(string)$s === '''.">!strlen($s)</weak_warning> === <weak_warning descr="Compare with an empty string instead: '(string)$u === '''.">!strlen($u)</weak_warning>;
    $r[] = <weak_warning descr="Compare with an empty string instead: '$t !== '''.">strlen($t) > 0</weak_warning> === $on;
    $r[] = $on == <weak_warning descr="Compare with an empty string instead: '$t === '''.">strlen($t) < 1</weak_warning>;
    $r[] = -<weak_warning descr="Compare with an empty string instead: '(string)$s === '''.">!strlen($s)</weak_warning> + 1;
    $r[] = <weak_warning descr="Compare with an empty string instead: '(string)$s === '''.">!strlen($s)</weak_warning> <=> $on;
    $r[] = <weak_warning descr="Compare with an empty string instead: '(string)$s !== '''.">strlen($s) != 0</weak_warning> & $on;
    $r[] = <weak_warning descr="Compare with an empty string instead: '(string)$s === '''.">!strlen($s)</weak_warning> ?? $on;
    $r[] = $on xor <weak_warning descr="Compare with an empty string instead: '(string)$s === '''.">!strlen($s)</weak_warning>;
    return $r;
}
echo json_encode([outer('abc', '', 'z', true), outer('', 'q', '', false)]), "\n";
