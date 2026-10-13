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
