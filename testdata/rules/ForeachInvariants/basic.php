<?php
function render(array $rows, $cap) {
    $total = count($rows);
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($n = 0; $n < $total; $n++) {
        print $rows[$n];
        echo $rows[$n]->title, " ({$rows[$n]}) ";
        $label = strtoupper($rows[$n]);
        $alias = &$rows[$n];
        if ($rows[$n] instanceof Stringable) {}
    }
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($n = 0; count($rows) > $n; ++$n) {
        $rows[$n][1] = 'x';
        echo $rows[$n][2];
    }
    <error descr="Replace the each() loop with foreach.">while</error> (list($slot, $entry) = each($rows)) {
        echo $entry;
    }
    for ($n = 0; $n < $cap; $n++) { echo $rows[$n]; }               // E5: parameter
    for ($n = 1; $n < count($rows); $n++) { echo $rows[$n]; }       // E2
    while (list($slot, $entry) = each($rows)) { unset($rows[$slot]); } // E7
}
