<?php
function tally(array $votes, $batch) {
    $flip  = array_flip(array_count_values($votes['names']));
    $pair  = count(array_count_values($votes['names'], $batch));
    $wrap  = count((array_count_values($votes['names'])));
    $expr  = count(array_count_values($votes['names']) + $batch);
    $obj   = $batch->count(array_count_values($votes['names']));
    $size  = sizeof(array_count_values($votes['names']));
    $none  = count(array_count_values());
    // spec divergence: extra arguments of the outer call
    $two   = array_keys(array_count_values($votes['names']), 2);
    $rec   = count(array_count_values($votes['names']), COUNT_RECURSIVE);
    return [$flip, $pair, $wrap, $expr, $obj, $size, $none, $two, $rec];
}
