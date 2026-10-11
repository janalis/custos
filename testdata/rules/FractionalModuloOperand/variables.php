<?php
function plain(int $n): int {
    $width = 1.5;
    return <error descr="Use fmod for fractional operands.">$n % $width</error>;
}
function valueCapture(int $n): int {
    $width = 1.5;
    $reader = function () use ($width) { $width = 3; return $width; };
    $reader();
    return <error descr="Use fmod for fractional operands.">$n % $width</error>;
}
function laterWriters(int $n, array $settings, string $query): int {
    $width = 1.5;
    $r = <error descr="Use fmod for fractional operands.">$n % $width</error>;
    $widen = function () use (&$width) { $width = 3; };
    $widen();
    extract($settings);
    parse_str($query, $width);
    return $r;
}
function explicitResult(int $n, string $query): int {
    $width = 1.5;
    parse_str($query, $parsed);
    $inner = function () use ($parsed) {
        return function (array $settings) use (&$parsed) { extract($settings); };
    };
    return <error descr="Use fmod for fractional operands.">$n % $width</error>;
}
function userCall(int $n): int {
    $width = 1.5;
    helper();
    $object = new Registry();
    $object->reset();
    Registry::reset();
    return <error descr="Use fmod for fractional operands.">$n % $width</error>;
}
function helper(): void {
}
