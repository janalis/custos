<?php
function layout(array $nodes, int $seed): array
{
    mt_srand($seed);
    $out = [];
    foreach ($nodes as $node) {
        $out[$node] = [mt_rand(0, 100), mt_rand(0, 100)];
    }
    $out['jitter'] = <warning descr="Prefer mt_rand() over rand().">rand(1, 6)</warning>;
    return $out;
}
