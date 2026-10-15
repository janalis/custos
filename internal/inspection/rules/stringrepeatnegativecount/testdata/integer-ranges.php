<?php
function negativeChoices(bool $pick): void {
    $n = $pick ? -2 : -7;
    <error descr="Supply a nonnegative repetition count.">str_repeat('x', $n)</error>;
}
function crossingChoices(bool $pick): void {
    $n = $pick ? -2 : 7;
    str_repeat('x', $n);
}
function decrementBody(): void {
    for ($n = -1; $n > -4; $n--) {
        <error descr="Supply a nonnegative repetition count.">str_repeat('x', $n)</error>;
    }
}
function onceBeforeRewrite($repeat): void {
    $n = -1;
    do {
        echo <error descr="Supply a nonnegative repetition count.">str_repeat('x', $n)</error>;
        $n = 5;
    } while ($repeat);
}
function optionalRewrite($change, $repeat): void {
    $n = -1;
    if ($change) { $n = 5; }
    do { str_repeat('x', $n); $n = 5; } while ($repeat);
}
function skippedEntry(): void {
    for ($n = 1; $n < 0; $n--) { str_repeat('x', $n); }
}
function overflowingCount(): void {
    $n = -9223372036854775807;
    $n -= 2;
    str_repeat('x', $n);
}
function unknownReference(): void {
    $n = -2;
    overwrite($n);
    str_repeat('x', $n);
}
