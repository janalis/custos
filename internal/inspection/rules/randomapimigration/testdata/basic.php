<?php
namespace Dice;

function throwDice($sides)
{
    <warning descr="Prefer mt_srand() over srand().">srand()</warning>;
    $a = <warning descr="Prefer random_int() over rand().">rand(1, $sides)</warning>;
    $b = <warning descr="Prefer random_int() over mt_rand().">\mt_rand(0, 9)</warning>;
    $c = <warning descr="Prefer mt_rand() over rand().">rand()</warning>;
    $d = mt_rand();
    $e = mt_rand(3);
    $f = random_int(1, 2);
}
