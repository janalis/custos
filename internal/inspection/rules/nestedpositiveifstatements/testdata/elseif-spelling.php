<?php
function spelling($a, $b, $c, $d) {
    // A chain written with elseif keeps that spelling.
    if ($a) {
        go(1);
    } elseif ($b) {
        go(2);
    } else {
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($c) {
            go(3);
        }
    }
    if ($a) {
        go(4);
    } else {
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($c) {
            go(5);
        } elseif ($d) {
            go(6);
        }
    }
    // So does a chain without elseif in a file that writes elseif.
    if ($a) {
        go(7);
    } else {
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($c) {
            go(8);
        }
    }
}
