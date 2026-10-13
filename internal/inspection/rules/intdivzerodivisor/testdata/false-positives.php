<?php

$top = 0;
$top++;
intdiv(12, $top);

function runningAverages(array $values): void
{
    $n = 0;
    foreach ($values as $value) {
        $n++;
        echo intdiv($value, $n), "\n";
    }
}

function steps(): void
{
    for ($d = 0; $d < 3; $d++) {
        if ($d) {
            echo intdiv(6, $d), "\n";
        }
    }
}

function attempts(): int
{
    $tries = 0;
    do {
        $tries++;
    } while ($tries < 3);

    return intdiv(6, $tries);
}

function preIncrement(): int
{
    $d = 0;
    ++$d;

    return intdiv(6, $d);
}

function decrement(): int
{
    $d = 0;
    $d--;

    return intdiv(6, $d);
}

function viaClosure(): int
{
    $d = 0;
    $set = function () use (&$d) {
        $d = 2;
    };
    $set();

    return intdiv(6, $d);
}

function steppedLate(): void
{
    $d = 0;
    while (more()) {
        if ($d) {
            echo intdiv(6, $d);
        }
        $d++;
    }
}

function pairs(array $rows): void
{
    $d = 0;
    foreach ($rows as [$n, $d]) {
        echo intdiv($n, $d);
    }
}

function keyed(array $rows): int
{
    $d = 0;
    foreach ($rows as $d => $row) {
    }

    return intdiv(6, $d);
}

function caught(): int
{
    $d = 0;
    try {
        risky();
    } catch (RuntimeException $d) {
    }

    return intdiv(6, $d);
}

function imported(array $row): int
{
    $d = 0;
    extract($row);

    return intdiv(6, $d);
}

$inline = function (array $values): int {
    $n = 0;
    foreach ($values as $value) {
        $n++;
    }

    return intdiv(6, $n);
};
