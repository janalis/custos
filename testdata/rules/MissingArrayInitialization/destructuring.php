<?php
function pairs($rows) {
    foreach ($rows as $r) {
        foreach ($r as $c) {
            $a[] = $c;
            $b[] = $c;
            <warning descr="Array '$d' is never initialised; initialise it before the loops.">$d[]</warning> = $c;
        }
    }
    foreach ($rows as [$a, $x]) {}
    [[$b]] = $rows;
    use_it([$d]);
}
