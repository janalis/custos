<?php
function afterExitCallable(string $value) {
    exit(...);
    $value = <warning descr="Assigning a value of type int does not match the parameter's declared type.">1</warning>;
}
function afterDieCallable(string $value) {
    die(...);
    $value = <warning descr="Assigning a value of type int does not match the parameter's declared type.">1</warning>;
}
function afterExitInvocation(string $value) {
    exit();
    $value = 1;
}
function afterImmediateExitInvocation(string $value) {
    ((exit(...))(1));
    $value = 1;
}
function afterNestedExitCallable(string $value) {
    (exit(...))(...);
    $value = <warning descr="Assigning a value of type int does not match the parameter's declared type.">1</warning>;
}
