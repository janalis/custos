<?php
function f($servers, $n) {
    for ($i = 0; $i <= $n; $i++) { // Try round-robin
    foreach ($servers as $server) {
        echo $server;
    }
    }
    if ($n) { # hash comment
        echo 1;
    }
    while ($n--) /* block */ {
        echo 2;
    }
}
function g($n) {
    if ($n) {
        echo 3; /* first */ // second
    }
    else {
        echo 4;
    } /* spans
        lines */
    if (!$n) {
        echo 5;
    } /* note */ echo 6;
}
