<?php
function tally(array $votes, $batch) {
    $names = array_values(array_unique($votes['names']));
    $total = \count(\array_unique(load($batch)));
    return [$names, $total];
}
