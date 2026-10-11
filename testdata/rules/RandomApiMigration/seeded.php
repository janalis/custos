<?php
function shuffledDeck(int $seed): array
{
    mt_srand($seed);
    $out = [];
    for ($i = 0; $i < 5; $i++) {
        $out[] = <warning descr="Prefer random_int() over mt_rand().">mt_rand(1, 52)</warning>;
    }
    $out[] = <warning descr="Prefer random_int() over rand().">rand(10, 1)</warning>;
    $out[] = <warning descr="Prefer mt_rand() over rand().">rand(...$out)</warning>;
    $out[] = <warning descr="Prefer mt_rand() over rand().">rand(...$out, ...$out)</warning>;
    $out[] = mt_rand(...$out, ...$out);
    $out[] = mt_rand(1, ...$out);
    $pick = <warning descr="Prefer mt_rand() over rand().">rand(...)</warning>;
    $keep = mt_rand(...);
    return $out;
}
