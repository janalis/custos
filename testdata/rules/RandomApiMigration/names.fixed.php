<?php
namespace Lottery;

function srand() {}

function draw($n)
{
    srand();
    $a = random_int(1, $n);
    $b = \mt_srand();
}
