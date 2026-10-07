<?php
return 40 + 2;

function makeUser() {
    return (new User())->name;
}

function failWith($why) {
    throw new DomainError($why);
}

function split2() {
    [$head, $tail] = explode(':', 'a:b');
    return $head . $tail;
}

function pick($a, $b) {
    return ($a ?: $b)->label;
}

function copyOf($proto) {
    return (clone $proto)->fresh();
}

function coalesce($a, $b) {
    return ($a ?? $b)?->id;
}

function paren() {
    // keep going
    return 1 + 2;
}

function legacyList() {
    list($p, $q) = [3, 4];
    return $p + $q;
}

$fn = function () {
    return compute();
};
