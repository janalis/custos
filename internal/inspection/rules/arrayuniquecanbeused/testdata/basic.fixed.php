<?php
function tally(array $votes, $batch) {
    $names = array_values(array_unique(['alpha', 'beta']));
    $total = \count(\array_unique([1, 2, 1]));
    return [$names, $total];
}
