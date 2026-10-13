<?php
function outerYoda($s, string $t, bool $on)
{
    $a = !<weak_warning descr="Compare with an empty string instead: ''' === (string)$s'.">!strlen($s)</weak_warning>;
    $b = <weak_warning descr="Compare with an empty string instead: ''' !== $t'.">strlen($t) > 0</weak_warning> === $on;
    $c = $on && <weak_warning descr="Compare with an empty string instead: ''' === $t'.">!strlen($t)</weak_warning>;
    return [$a, $b, $c];
}
echo json_encode([outerYoda('abc', '', true), outerYoda('', 'q', false)]), "\n";
