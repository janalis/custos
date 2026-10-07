<?php
namespace Lottery;

function srand() {}

function draw($n)
{
    srand();
    $a = <warning descr="Prefer random_int() over rand().">RAND(1, $n)</warning>;
    $b = <warning descr="Prefer mt_srand() over srand().">\SRand()</warning>;
}
