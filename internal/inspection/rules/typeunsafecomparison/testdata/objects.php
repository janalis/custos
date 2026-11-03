<?php
// == on two objects compares their properties, === their identity.
final class Money
{
    public function __construct(public int $cents) {}
}

function same(Money $a, Money $b, object $c, int $n)
{
    return [
        $a == $b,
        $a != $c,
        <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$a == $n</weak_warning>,
    ];
}

function sameOrNull(?Money $a, Money|null $b, ?int $n, $x)
{
    return [
        $a != $b,
        <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$a == $n</weak_warning>,
        <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$a == $x</weak_warning>,
    ];
}
