<?php
function f($s) {
    $s = str_replace('a', 'b', $s);
    $s = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace(['c', 'd'], 'e', $s)</warning>;
    return $s;
}
