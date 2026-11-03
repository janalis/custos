<?php
function compare(array $a, array $b, mixed $m, $u, string $s, int|array $mixedType, ?array $maybe)
{
    $same = $a == $b;
    $fromMixed = $m == $a;
    $unknown = $a != $u;
    $list = [1, 2] == $a;
    $text = <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$a == $s</weak_warning>;
    $either = <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$mixedType == $a</weak_warning>;
    $nullable = <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$maybe == $b</weak_warning>;
    return [$same, $fromMixed, $unknown, $list, $text, $either, $nullable];
}

// An object never equals an array, nor do false and null a non-empty one.
function unions(array|false $r, array|\stdClass $o, ?array $n, array|bool $b, array $a)
{
    return [
        $r == ['b' => 1.0],
        ([1]) != $n,
        $o == $a,
        <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$r == []</weak_warning>,
        <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$r == [...$a]</weak_warning>,
        <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$n == $a</weak_warning>,
        <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$b == ['x']</weak_warning>,
    ];
}
