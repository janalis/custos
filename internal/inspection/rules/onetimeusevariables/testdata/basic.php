<?php
<warning descr="Variable $total is used only once; inline its value.">$total</warning> = 40 + 2;
return $total;

function makeUser() {
    <warning descr="Variable $user is used only once; inline its value.">$user</warning> = new User();
    return $user->name;
}

function failWith($why) {
    <warning descr="Variable $err is used only once; inline its value.">$err</warning> = new DomainError($why);
    throw $err;
}

function split2() {
    <warning descr="Variable $parts is used only once; inline its value.">$parts</warning> = explode(':', 'a:b');
    [$head, $tail] = $parts;
    return $head . $tail;
}

function pick($a, $b) {
    <warning descr="Variable $chosen is used only once; inline its value.">$chosen</warning> = $a ?: $b;
    return $chosen->label;
}

function copyOf($proto) {
    <warning descr="Variable $dup is used only once; inline its value.">$dup</warning> = clone $proto;
    return $dup->fresh();
}

function coalesce($a, $b) {
    <warning descr="Variable $o is used only once; inline its value.">$o</warning> = $a ?? $b;
    return $o?->id;
}

function paren() {
    /** describes the sum */
    <warning descr="Variable $m is used only once; inline its value.">$m</warning> = (1 + 2);
    // keep going
    return $m;
}

function legacyList() {
    <warning descr="Variable $pair is used only once; inline its value.">$pair</warning> = [3, 4];
    list($p, $q) = $pair;
    return $p + $q;
}

$fn = function () {
    <warning descr="Variable $r is used only once; inline its value.">$r</warning> = compute();
    return $r;
};

function namelessNotVar($auth) {
    /** @return \App\Guard */
    <warning descr="Variable $guard is used only once; inline its value.">$guard</warning> = $auth->guard('api');
    return $guard;
}
