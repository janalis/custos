<?php
namespace Stats;

function distinct(array $tags, array $words) {
    $a = <weak_warning descr="Use 'count(array_unique($tags))' instead (array_unique() is fast since PHP 7.2).">Count(ARRAY_COUNT_VALUES($tags))</weak_warning>;
    $b = <weak_warning descr="Use 'array_values(array_unique($words))' instead (array_unique() is fast since PHP 7.2).">\Array_Keys(Array_Count_Values($words))</weak_warning>;
    $c = <weak_warning descr="Use 'count(array_unique($words))' instead (array_unique() is fast since PHP 7.2).">count(\array_count_values($words))</weak_warning>;
    $d = SIZEOF(ARRAY_COUNT_VALUES($tags));
    return [$a, $b, $c, $d];
}
