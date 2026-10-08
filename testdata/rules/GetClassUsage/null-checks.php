<?php
// Earlier boolean uses of the argument count as null checks (C3, C4), even
// in unrelated branches.
function viaIsset(?Order $o) {
    $set = isset($o);
    return get_class($o);
}
function viaEmpty(?Order $o) {
    $none = empty($o);
    return get_class($o);
}
function viaInstanceof(?Order $o) {
    $is = $o instanceof Order;
    return get_class($o);
}
function viaIf(?Order $o) {
    if ($o) { echo 1; }
    return get_class($o);
}
function viaElseIf(?Order $o, $flag) {
    if ($flag) { echo 1; } elseif ($o) { echo 2; }
    return get_class($o);
}
function viaWhile(?Order $o) {
    while ($o) { break; }
    return get_class($o);
}
function viaDoWhile(?Order $o) {
    do { echo 1; } while ($o);
    return get_class($o);
}
function viaNot(?Order $o) {
    $missing = !$o;
    return get_class($o);
}
function viaAnd(?Order $o, $flag) {
    $both = $flag && $o;
    return get_class($o);
}
function viaTernary(?Order $o) {
    $label = $o ? 'set' : 'unset';
    return get_class($o);
}
function viaNullOnLeft(?Order $o) {
    $same = null === $o;
    return get_class($o);
}
// Not checks: short ternary, other operators, comparison with non-null.
function viaElvis(?Order $o) {
    $label = $o ?: 'unset';
    return <warning descr="get_class() rejects null on PHP 7.2+; guard the argument.">get_class($o)</warning>;
}
function viaPlus(?Order $o) {
    $sum = $o . 'x';
    return <warning descr="get_class() rejects null on PHP 7.2+; guard the argument.">get_class($o)</warning>;
}
function viaFalse(?Order $o) {
    $same = $o === false;
    return <warning descr="get_class() rejects null on PHP 7.2+; guard the argument.">get_class($o)</warning>;
}
// Arrow functions: no null-check search (reported when the type has null);
// a null default alone (D4b) does not apply.
$arrow = fn(?Order $o) => <warning descr="get_class() rejects null on PHP 7.2+; guard the argument.">get_class($o)</warning>;
$untyped = fn($o) => get_class($o);
// Top level: an untyped variable is not a null-defaulted parameter.
get_class($unknown);
strlen($unknown);
$fn($unknown);
