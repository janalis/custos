<?php
function sameOperands($a) {
    if (<error descr="Both operands are the same.">strlen($a) === STRLEN($a)</error>) {}
    if (<error descr="Both operands are the same.">count($a) > COUNT($a)</error>) {}
    if (<error descr="Both operands are the same.">Box::K == box::K</error>) {}
    if ($a->size === $a->Size) {}
}

// Operands that may evaluate differently are not the same.
function varying($a, Repo $r) {
    return [
        mt_rand() == mt_rand(),
        $a->next() !== $a->next(),
        Repo::find(1) === Repo::find(1),
        new Repo() == new Repo(),
        userRandom() == userRandom(),
        $a++ == $a++,
    ];
}
