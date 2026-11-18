<?php
function bump(array $m, int $b, $repo)
{
    $m[$b++] = $m[$b++] + 1;        // the target is evaluated twice
    $m[$b = 2] = $m[$b = 2] . 'x';
    $f = [$m[fn () => $b++] .= "y"]; // closures are not run here
    $m[$b] = $m[$b] += 1;
    return $m;
}
