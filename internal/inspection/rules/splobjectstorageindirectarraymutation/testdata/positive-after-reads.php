<?php

function countedFirst(): int
{
    $s = new SplObjectStorage();
    $o = new stdClass();
    $s[$o] = ["c" => 1];
    $n = $s->count();
    <warning descr="Assign the modified storage value back.">$s[$o]["c"] = 2</warning>;
    return $n;
}

function arrayRestoredAfterObject(): void
{
    $s = new SplObjectStorage();
    $o = new stdClass();
    $s[$o] = ["c" => 1];
    $s->attach($o, new ArrayObject());
    $s[$o] = ["c" => 3];
    <warning descr="Assign the modified storage value back.">$s[$o]["c"]++</warning>;
}

function anotherKeyHoldsAnArray(object $p): void
{
    $s = new SplObjectStorage();
    $o = new stdClass();
    $s[$o] = ["c" => 1];
    $s[$p] = ["d" => 1];
    $unrelated = function () use ($p) {
        return $p;
    };
    $t = [];
    $t[0] = $unrelated;
    unset($t[0]);
    <warning descr="Assign the modified storage value back.">$s[$o]["c"] = 2</warning>;
}
