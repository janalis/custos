<?php

/** @param float|null $ratio */
function probe($ratio, ?\DateTime $when, array $rows, ?string $s, int|bool|null $x, ?array $a)
{
    return [
        empty($ratio),
        <weak_warning descr="Replace with 'null !== $when'.">!empty($when)</weak_warning>,
        static function () use ($when) { return<weak_warning descr="Replace with 'null !== $when'.">!empty($when)</weak_warning>; },
        empty($rows),
        empty($s),
        empty($x),
        empty($a),
        empty($unknown),
    ];
}
