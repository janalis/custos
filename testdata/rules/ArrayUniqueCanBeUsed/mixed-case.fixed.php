<?php
namespace Stats;

function distinct(array $tags, array $words) {
    $a = count(array_unique($tags));
    $b = \array_values(array_unique($words));
    $c = count(\array_unique($words));
    $d = SIZEOF(ARRAY_COUNT_VALUES($tags));
    return [$a, $b, $c, $d];
}
