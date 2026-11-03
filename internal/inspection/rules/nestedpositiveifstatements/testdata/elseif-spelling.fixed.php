<?php
function spelling($a, $b, $c, $d) {
    // A chain written with elseif keeps that spelling.
    if ($a) {
        go(1);
    } elseif ($b) {
        go(2);
    } elseif ($c) {
        go(3);
    }
    if ($a) {
        go(4);
    } elseif ($c) {
        go(5);
    } elseif ($d) {
        go(6);
    }
    if ($a) {
        go(7);
    } else if ($c) {
        go(8);
    }
}
