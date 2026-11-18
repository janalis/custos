<?php
str_repeat('cd', 3);

function render(array $nodes): void
{
    $depth = -1;
    foreach ($nodes as $node) {
        $depth++;
        echo str_repeat('  ', $depth), $node, "\n";
    }
}

function once(): string
{
    $n = -1;
    ++$n;
    return str_repeat('ab', $n) . '|';
}

function preIncrementLoop(): void
{
    $i = -1;
    while (++$i < 3) {
        echo str_repeat('a', $i);
    }
}

function guardedFor(): void
{
    for ($j = -1; $j < 3; $j++) {
        if ($j < 0) {
            continue;
        }
        echo str_repeat('a', $j);
    }
}

function steppedAtLoopEnd(array $rows): void
{
    $level = -1;
    foreach ($rows as $row) {
        if ($level >= 0) {
            echo str_repeat(' ', $level), $row;
        }
        $level++;
    }
}

function capturedByReference(): string
{
    $w = -1;
    $set = function () use (&$w) { $w = 2; };
    $set();
    return str_repeat('a', $w);
}

$top = -1;
$top++;
echo str_repeat('a', $top);

function steppedCopy(): string
{
    $n = -1;
    $n++;
    $m = $n;
    return str_repeat('a', $m);
}

function rebound(array $vars): string
{
    $n = -1;
    extract($vars);
    return str_repeat('a', $n);
}

function longChain(): string
{
    $a = -1;
    $b = $a;
    $c = $b;
    $d = $c;
    $e = $d;
    $f = $e;
    $g = $f;
    $h = $g;
    $i = $h;
    $j = $i;
    return <error descr="Supply a nonnegative repetition count.">str_repeat('a', $j)</error>;
}
