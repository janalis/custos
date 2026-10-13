<?php
function show(array $items) {
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < COUNT($items); $i++) {
        echo $items[$i];
    }
    <error descr="Replace the each() loop with foreach.">while</error> (list($k, $v) = Each($items)) {
        echo $v;
    }
}
