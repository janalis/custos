<?php
define('DICE_FACES', 6);

class Lottery { const LOW = 1; const HIGH = 49; }

function roll($bonus = 10, $other = null)
{
    $floor = 100;
    return [
        <error descr="Minimum is greater than maximum in this random range.">mt_rand(9, 3)</error>,
        mt_rand(3, 9),
        <error descr="Minimum is greater than maximum in this random range.">rand(DICE_FACES, 1)</error>,
        rand(1, DICE_FACES),
        <error descr="Minimum is greater than maximum in this random range.">random_int(Lottery::HIGH, Lottery::LOW)</error>,
        <error descr="Minimum is greater than maximum in this random range.">\random_int($floor, -1)</error>,
        <error descr="Minimum is greater than maximum in this random range.">random_int($bonus, 0)</error>,
        <error descr="Minimum is greater than maximum in this random range.">random_int(PHP_INT_MAX, PHP_INT_MIN)</error>,
        random_int(0, PHP_INT_MAX),
        <error descr="Minimum is greater than maximum in this random range.">random_int(0, PHP_INT_MIN)</error>,
        random_int(7, 7),
        <error descr="Minimum is greater than maximum in this random range.">random_int(0x10, 1)</error>,
        <error descr="Minimum is greater than maximum in this random range.">random_int(2.5, 1)</error>,
        <error descr="Minimum is greater than maximum in this random range.">random_int(1_000, 1)</error>,
        random_int(5, count([1])),
        random_int(10 - 1, 1),
        random_int($other ? 9 : 8, 1),
        <error descr="Minimum is greater than maximum in this random range.">random_int(- 5, -6)</error>,
        mt_rand(5),
        rand(),
    ];
}

$top = 10;
random_int($top, 1);
