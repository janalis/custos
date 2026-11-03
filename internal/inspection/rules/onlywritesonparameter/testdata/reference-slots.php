<?php

// An element may be a reference: writing it writes what it refers to.
function literal(): int
{
    $x = 123;
    $slots = [&$x];
    $slots[0] = 999;
    return $x;
}

function assigned(array $src): int
{
    $y = 1;
    $map = [];
    $map['k'] = &$y;
    $copy = $map;
    $copy['k'] = 2;
    return $y;
}

function looped(array $rows): array
{
    foreach ($rows as &$row) {
        $row = trim($row);
    }
    unset($row);
    $view = $rows;
    $view[0] = 'x';
    return $rows;
}

// Without references, the element write is lost.
function plain(): void
{
    $a = [1];
    <weak_warning descr="Value is only written here and never read; the write is lost.">$a[0]</weak_warning> = 2;
}

function mixed_writes(): int
{
    $x = 1;
    $c = [&$x];
    $c = [3];
    $c[0] = 4;
    return $x;
}
