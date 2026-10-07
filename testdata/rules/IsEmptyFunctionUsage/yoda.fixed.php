<?php

/** @param float|null $ratio */
function probe($ratio, ?\DateTime $when, array $rows, ?string $s, int|bool|null $x, ?array $a)
{
    return [
        empty($ratio),
        null !== $when,
        empty($rows),
        empty($s),
        empty($x),
        empty($a),
        empty($unknown),
    ];
}
