<?php
function sameLabel(Collator $c, string $left, string $right): array {
    return [
        <warning descr="Compare collation results explicitly.">$c->compare($left, $right) == true</warning>,
        <warning descr="Compare collation results explicitly.">true == $c->compare($left, $right)</warning>,
        <warning descr="Compare collation results explicitly.">true === $c->compare($left, $right)</warning>,
        <warning descr="Compare collation results explicitly.">TRUE === ($c->compare($left, $right))</warning>,
        false === $c->compare($left, $right),
        0 === $c->compare($left, $right),
        true !== $c->compare($left, $right),
    ];
}
