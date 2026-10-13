<?php
function destructured(array $pairs)
{
    foreach ($pairs as $k => [$x, $y]) {
        $pairs[$k] = $x + $y;
    }
    return $pairs;
}
