<?php
function capturedAfter($value) {
    <error descr="Parameter is overwritten before its value is used.">$value</error> = 1;
    $capture = fn() => $value;
    return [$value, $capture];
}
function capturedBefore($value) {
    $capture = fn() => $value;
    $value = 1;
    return [$value, $capture];
}
function reflectedAfter($value) {
    <error descr="Parameter is overwritten before its value is used.">$value</error> = 1;
    return [$value, get_defined_vars()];
}
function compactedOther($value, $other) {
    $snapshot = compact('other');
    <error descr="Parameter is overwritten before its value is used.">$value</error> = 1;
    return [$value, $snapshot];
}
function arrowShadow($value) {
    $capture = fn($value) => $value;
    <error descr="Parameter is overwritten before its value is used.">$value</error> = 1;
    return [$value, $capture];
}
function getterAfter($value) {
    <error descr="Parameter is overwritten before its value is used.">$value</error> = 1;
    return [$value, func_get_args()];
}
function getterOther($value, $other) {
    $saved = func_get_arg(1);
    <error descr="Parameter is overwritten before its value is used.">$value</error> = 1;
    return [$value, $saved];
}
