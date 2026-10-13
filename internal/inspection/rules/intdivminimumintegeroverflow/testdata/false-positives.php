<?php

function incrementedDividend(): int
{
    $min = PHP_INT_MIN;
    $min++;

    return intdiv($min, -1);
}

function decrementedDivisor(): int
{
    $min = PHP_INT_MIN;
    $divisor = -1;
    $divisor--;

    return intdiv($min, $divisor);
}

function compoundDivisor(): int
{
    $divisor = -1;
    $divisor -= 1;

    return intdiv(PHP_INT_MIN, $divisor);
}

function destructuredDividend(array $row): int
{
    $min = PHP_INT_MIN;
    [$min] = $row;

    return intdiv($min, -1);
}

function iteratedDividend(): array
{
    $min = PHP_INT_MIN;
    $out = [];
    foreach ([1, 2] as $min) {
        $out[] = intdiv($min, -1);
    }

    return $out;
}

function scannedDividend(string $text): int
{
    $min = PHP_INT_MIN;
    sscanf($text, '%d', $min);

    return intdiv($min, -1);
}

function scannedVariadicTail(string $text): int
{
    $min = PHP_INT_MIN;
    sscanf($text, '%d %d', $first, $min);

    return intdiv($min, -1) + $first;
}

function matchedDivisor(): int
{
    $divisor = -1;
    preg_match(subject: 'x', pattern: '/x/', matches: $divisor);

    return intdiv(PHP_INT_MIN, $divisor);
}

function copiedFromChangedSource(): int
{
    $source = PHP_INT_MIN;
    $source++;
    $min = $source;

    return intdiv($min, -1);
}

function changedLaterInLoop(): void
{
    $min = PHP_INT_MIN;
    for ($i = 0; $i < 2; $i++) {
        echo intdiv($min, -1);
        $min++;
    }
}

function aliasedDividend(): int
{
    $min = PHP_INT_MIN;
    $alias = &$min;
    $alias = 3;

    return intdiv($min, -1);
}

function referencedInArray(): int
{
    $min = PHP_INT_MIN;
    $box = [&$min];
    $box[0] = 3;

    return intdiv($min, -1);
}

function capturedByReference(): int
{
    $min = PHP_INT_MIN;
    $reset = function () use (&$min) {
        $min = 1;
    };
    $reset();

    return intdiv($min, -1);
}

function passedToUnknownCallee($object): int
{
    $min = PHP_INT_MIN;
    $object->adjust($min);

    return intdiv($min, -1);
}

function unpackedIntoCall(): int
{
    $divisor = -1;
    var_dump(...$divisor);

    return intdiv(PHP_INT_MIN, $divisor);
}

function unsetDividend(): int
{
    $min = PHP_INT_MIN;
    unset($min);

    return intdiv($min, -1);
}

function otherDivisor(): int
{
    return intdiv(PHP_INT_MIN, -2) + intdiv(5, -1);
}
