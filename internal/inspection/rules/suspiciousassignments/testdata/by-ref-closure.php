<?php
// A closure importing $captured by reference may fill it when called.
function capture(callable $register) {
    $captured = null;
    $register(function (string $sql, array $args) use (&$captured) {
        $captured = [$sql, $args];
    });
    [$sql, $args] = $captured;

    $plain = null;
    $register(function () use ($plain) {});
    <error descr="Destructuring a value that is not an array.">[$x] = $plain</error>;
}

// A parameter a by-reference closure may replace is not proven unread.
function reloaded($value) {
    $value = load_value();
    $refresh = function () use (&$value) { $value = load_value(); };
    $refresh();
    return $value;
}
