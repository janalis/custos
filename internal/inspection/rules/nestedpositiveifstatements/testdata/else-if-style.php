<?php
// A file that writes `else if` more often than `elseif` keeps `else if`
// for a chain without elseif.
function style($a, $b, $c) {
    if ($a) {
        go(1);
    } else if ($b) {
        go(2);
    }
    if ($b) {
        go(3);
    } else if ($c) {
        go(4);
    }
    if ($c) {
        go(5);
    } elseif ($a) {
        go(6);
    }
    if ($a) {
        go(7);
    } else {
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($c) {
            go(8);
        }
    }
}
