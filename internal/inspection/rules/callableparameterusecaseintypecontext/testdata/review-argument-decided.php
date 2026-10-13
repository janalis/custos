<?php
// Built-ins whose result type is decided by their arguments.
function square(float $side, float $f, int $n, string $s, $any): array
{
    $side = pow($side, 2);
    $f = pow(2, 3);
    $n = <warning descr="Assigning a value of type float does not match the parameter's declared type.">pow(2, 0.5)</warning>;
    $s = <warning descr="Assigning a value of type int does not match the parameter's declared type.">pow(2, 3)</warning>;
    $n = pow($n, 2);
    $s = pow($any, 2);
    $s = pow(num: 2, exponent: 3);
    return [$side, $f, $n, $s];
}

function clocks(int $since, float $f, array $parts, string $s, int $n, $any): array
{
    $since = hrtime(true);
    $f = hrtime(true);
    $parts = hrtime();
    $parts = hrtime(false);
    $since = <warning descr="Assigning a value of type array does not match the parameter's declared type.">hrtime()</warning>;
    $since = hrtime($any);
    $s = microtime();
    $s = microtime(false);
    $f = microtime(true);
    $n = <warning descr="Assigning a value of type string does not match the parameter's declared type.">microtime()</warning>;
    return [$since, $f, $parts, $s, $n];
}

function environment(string $dir, array $all, string $name, $any): array
{
    $dir = getenv('HOME');
    $dir = getenv($name);
    $all = getenv();
    $all = getenv(null);
    $dir = <warning descr="Assigning a value of type array does not match the parameter's declared type.">getenv()</warning>;
    $all = <warning descr="Assigning a value of type string does not match the parameter's declared type.">getenv('HOME')</warning>;
    $dir = getenv($any);
    return [$dir, $all];
}

function versions(bool $old, int $order, string $op, $any): array
{
    $old = version_compare(PHP_VERSION, '8.0', '<');
    $old = version_compare(PHP_VERSION, '8.0', $op);
    $order = version_compare(PHP_VERSION, '8.0');
    $old = <warning descr="Assigning a value of type int does not match the parameter's declared type.">version_compare(PHP_VERSION, '8.0')</warning>;
    $order = <warning descr="Assigning a value of type bool does not match the parameter's declared type.">version_compare(PHP_VERSION, '8.0', '<')</warning>;
    $old = version_compare(PHP_VERSION, '8.0', $any);
    return [$old, $order];
}
