<?php
function sweep(): void
{
    for ($x = -2; $x <= 2; $x++) {
        if (abs($x) <= 1) {
            echo asin($x), "\n";
        }
    }
}
function clamp(): float
{
    $x = 3;
    while ($x > 1) {
        $x--;
    }
    return acos($x);
}
function step(): float
{
    $x = 2;
    $x--;
    return asin($x);
}
function steppedAfterUseInLoop(): void
{
    $x = 2;
    foreach ([1, 2] as $_) {
        $x--;
        echo asin($x);
    }
    $y = 1;
    do {
        echo acos($y);
        $y++;
    } while ($y < 1);
}
function capturedByReference(): float
{
    $x = 2;
    $reset = function () use (&$x) {
        $x = 0;
    };
    $reset();
    return asin($x);
}
function extracted(array $overrides): float
{
    $x = 2;
    extract($overrides);
    return asin($x);
}
$top = 2;
$top--;
echo asin($top);
