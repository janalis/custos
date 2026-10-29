<?php
function shuffleDeck(array $cards): array
{
    srand();
    $top = getrandmax();
    $cut = rand();
    $pick = <warning descr="Prefer random_int() over rand().">rand(0, count($cards) - 1)</warning>;
    return [$top, $cut, $cards[$pick]];
}

function roll(int $sides): int
{
    return <warning descr="Prefer random_int() over rand().">rand(1, $sides)</warning>;
}
