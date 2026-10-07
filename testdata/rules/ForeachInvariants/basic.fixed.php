<?php
function render(array $rows, $cap) {
    foreach ($rows as $n => $nValue) {
        print $rows[$n];
        echo $nValue->title, " ($nValue) ";
        $label = strtoupper($nValue);
        $alias = &$rows[$n];
        if ($nValue instanceof Stringable) {}
    }
    foreach ($rows as $n => $nValue) {
        $rows[$n][1] = 'x';
        echo $nValue[2];
    }
    foreach ($rows as $entry) {
        echo $entry;
    }
    for ($n = 0; $n < $cap; $n++) { echo $rows[$n]; }               // E5: parameter
    for ($n = 1; $n < count($rows); $n++) { echo $rows[$n]; }       // E2
    while (list($slot, $entry) = each($rows)) { unset($rows[$slot]); } // E7
}
