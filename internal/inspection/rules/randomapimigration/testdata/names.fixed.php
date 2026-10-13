<?php
namespace Lottery;

function srand() {}

function draw($n)
{
    srand();
    $a = RAND(1, $n);
    $b = \mt_srand();
}
