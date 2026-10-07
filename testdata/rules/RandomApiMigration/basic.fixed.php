<?php
namespace Dice;

function throwDice($sides)
{
    mt_srand();
    $a = random_int(1, $sides);
    $b = \random_int(0, 9);
    $c = mt_rand();
    $d = mt_rand();
    $e = mt_rand(3);
    $f = random_int(1, 2);
}
