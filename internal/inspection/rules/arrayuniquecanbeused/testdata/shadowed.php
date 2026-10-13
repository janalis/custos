<?php
namespace Tally;

function array_count_values(array $a) { return $a; }
function count($a) { return 1; }

function sizes(array $votes) {
    $a = count(\array_count_values($votes));
    $b = \count(array_count_values($votes));
    $c = \array_keys(\Other\array_count_values($votes));
    return [$a, $b, $c];
}
