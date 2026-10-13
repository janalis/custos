<?php
function totals(array $rows, $tax) {
    $total = array_sum($rows);
    $make = fn($extra) => compact('total', 'tax', 'extra', <error descr="Variable '$later' may be undefined when compact() runs.">'later'</error>);
    $nested = fn() => fn() => compact('rows', <error descr="Variable '$nope' may be undefined when compact() runs.">'nope'</error>);
    $later = 1;
    return [$make, $nested];
}

$top = fn($a) => compact('a', 'b');
