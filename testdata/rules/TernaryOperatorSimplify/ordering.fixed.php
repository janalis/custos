<?php
function ordering(int $count, string $name, float $ratio, $any, ?int $max) {
    $r = $count >= 10;
    $r = $name > 'm';
    $r = $max < 3;
    $r = !($ratio < 0.5);
    $r = !($any > $count);
    $r = $ratio != 0.5;
    $r = $ratio < 0.5;
    return $r;
}
