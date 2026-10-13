<?php
function pick(array $xs)
{
    $n = 0;
    foreach ($xs as $x) {
        ++$n;
    }
    $lo = 10;
    $lo -= 9;
    return [mt_rand(1, $n), rand($lo, 5)];
}
