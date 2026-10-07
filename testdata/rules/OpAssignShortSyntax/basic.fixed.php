<?php
function tally(int $count, string $label, array $grid, $buf) {
    $count += 5;
    $count -= $step;
    $count %= 7;
    $count <<= 1;
    $label .= '-' . $count . '!';
    $count *= 3 * $k;
    $count += ($k * 2);
    $grid[1] |= 4;
    $buf[0] .= 'z';
    $this->n ^= $m;
}
