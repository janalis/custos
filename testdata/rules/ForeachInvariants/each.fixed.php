<?php
function eachLoops(array $a, array $b) {
    while (list($k, $v) = each()) { echo $v; }                  // no argument
    while (list($k, $v) = each($a, $b)) { echo $v; }            // two arguments
    list($k, $v) = each($a);                                     // not a loop condition
    while (list($k, $v) = each($a)) {}                           // empty body
    while (list($k, $v) = each($a)) { /* nothing */ }            // comment only
    while ($k = each($a)) { echo $k; }                           // not a destructuring
    while (list($k, $v) = next($a)) { echo $v; }                 // other function
    for (; list($k, $v) = each($a);) {
        echo $v;
    }
    foreach ($b as $k => $v) {
        echo $k, $v;
    }
    while ([, $v] = each($b)) {
        echo $v;
    }
    while (list($k, $v, $w) = each($b)) {
        echo $v;
    }
    while (list($k, $o->v) = each($b)) {
        echo $k;
    }
}
