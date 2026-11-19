<?php

class Money extends GMP {}

class Orphan extends MissingBase {}

function converted(SimpleXMLElement $xml)
{
    return $xml * 2;
}

/** @param array $documented */
function allowed(array $a, array $b, ?array $maybe, array|int $either, mixed $anything, iterable $items, $unknown, $documented, int $n, string $s, GMP $gmp, Money $money, Orphan $orphan, Missing $missing, object $thing, array|string $text)
{
    $merged = $a + $b;
    $a += $b;
    $x = $a + $maybe;
    $x = $a + $anything;
    $x = $a + $items;
    $x = $a + $unknown;
    $x = $a + $either;
    $x = $maybe - 1;
    $x = $either * 2;
    $x = $anything * 2;
    $x = $documented * 2;
    $x = $n + 1;
    $x = $s . 'text';
    $x = $gmp + 1;
    $x = $money * 2;
    $x = $orphan + 1;
    $x = $missing + 1;
    $x = $thing + 1;
    $x = $text - 1;
    $n = 5;
    $x = $n < 3;
}
