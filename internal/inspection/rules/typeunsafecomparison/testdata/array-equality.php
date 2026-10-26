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
