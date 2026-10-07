<?php
function eachLoops(array $a, array $b) {
    while (list($k, $v) = each()) { echo $v; }                  // no argument
    while (list($k, $v) = each($a, $b)) { echo $v; }            // two arguments
    list($k, $v) = each($a);                                     // not a loop condition
    while (list($k, $v) = each($a)) {}                           // empty body
    while (list($k, $v) = each($a)) { /* nothing */ }            // comment only
    while ($k = each($a)) { echo $k; }                           // not a destructuring
    while (list($k, $v) = next($a)) { echo $v; }                 // other function
    <error descr="Replace the each() loop with foreach.">for</error> (; list($k, $v) = each($a);) {
        echo $v;
    }
    <error descr="Replace the each() loop with foreach.">while</error> (list($k, $v) = each($b)) {
        echo $k, $v;
    }
    <error descr="Replace the each() loop with foreach.">while</error> ([, $v] = each($b)) {
        echo $v;
    }
    <error descr="Replace the each() loop with foreach.">while</error> (list($k, $v, $w) = each($b)) {
        echo $v;
    }
    <error descr="Replace the each() loop with foreach.">while</error> (list($k, $o->v) = each($b)) {
        echo $k;
    }
}
