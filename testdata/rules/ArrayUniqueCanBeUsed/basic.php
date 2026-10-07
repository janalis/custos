<?php
function tally(array $votes, $batch) {
    $names = <weak_warning descr="Use 'array_values(array_unique($votes['names']))' instead (array_unique() is fast since PHP 7.2).">array_keys(array_count_values($votes['names']))</weak_warning>;
    $total = <weak_warning descr="Use 'count(array_unique(load($batch)))' instead (array_unique() is fast since PHP 7.2).">\count(\array_count_values(load($batch)))</weak_warning>;
    return [$names, $total];
}
