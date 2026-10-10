<?php
function tally(array $votes, $batch) {
    $names = <weak_warning descr="Use 'array_values(array_unique(['alpha', 'beta']))' instead (array_unique() is fast since PHP 7.2).">array_keys(array_count_values(['alpha', 'beta']))</weak_warning>;
    $total = <weak_warning descr="Use 'count(array_unique([1, 2, 1]))' instead (array_unique() is fast since PHP 7.2).">\count(\array_count_values([1, 2, 1]))</weak_warning>;
    return [$names, $total];
}
