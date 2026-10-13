<?php
function tally(array $votes, $batch) {
    $names = array_keys(array_count_values($votes['names']));
    $total = \count(\array_count_values(load($batch)));
    return [$names, $total];
}
