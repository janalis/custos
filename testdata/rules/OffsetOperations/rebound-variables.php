<?php
// A variable rebound by destructuring or through a reference holds a value
// the inferred type does not describe: no report.
function destructured(): string
{
    $n = 0;
    [$n, $m] = [['x'], 2];
    return $n[0] . $m;
}

function listed(int $n): string
{
    list('a' => $n) = ['a' => 'str'];
    return $n[0];
}

function converted(float $n)
{
    settype($n, 'array');
    return $n[0];
}

function convertedByName(float $n)
{
    settype(type: 'array', var: $n);
    return $n[0];
}

function scanned(string $line)
{
    $first = 0;
    $second = 0;
    sscanf($line, '%s %s', $first, $second);
    return $first[0] . $second[0];
}

function aliased(int $n)
{
    $r = &$n;
    $r = [1];
    return $n[0];
}

function bound(array $source)
{
    $n = 0;
    $n = &$source;
    return $n[0];
}

function captured()
{
    $n = 0;
    $fill = function () use (&$n) {
        $n = [1];
    };
    $fill();
    return ($n)[0];
}

function callTime()
{
    $n = 0;
    fillLegacy(&$n);
    return $n[0];
}

function indexAfterConversion(array $rows)
{
    $s = 'abc';
    settype($s, 'object');
    return $s[array_keys($rows)];
}

// Still reported: the variable is only read, copied or passed by value.
function passedByValue(int $n)
{
    $copy = $n;
    $twice = function () use ($n) {
        return $n * 2;
    };
    echo abs($n), unknownHelper($n), $copy, $twice();
    return <error descr="'$n' does not support offset access (types: int).">$n[0]</error>;
}

function namedByValue(int $n)
{
    preg_match(pattern: '/x/', subject: $n);
    return <error descr="'$n' does not support offset access (types: int).">$n[0]</error>;
}

function notAVariable(array $rows)
{
    return <error descr="'count($rows)' does not support offset access (types: int).">count($rows)[0]</error>;
}
