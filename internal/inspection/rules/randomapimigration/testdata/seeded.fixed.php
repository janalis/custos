<?php
function shuffledDeck(int $seed): array
{
    mt_srand($seed);
    $out = [];
    for ($i = 0; $i < 5; $i++) {
        $out[] = mt_rand(1, 52);
    }
    $out[] = rand(10, 1);
    $out[] = mt_rand(...$out);
    $out[] = mt_rand(...$out, ...$out);
    $out[] = mt_rand(...$out, ...$out);
    $out[] = mt_rand(1, ...$out);
    $pick = mt_rand(...);
    $keep = mt_rand(...);
    return $out;
}
