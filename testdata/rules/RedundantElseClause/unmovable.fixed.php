<?php
// The clause's code cannot move: the if is an unbraced body, or the clause
// declares a function (PHP would hoist it).
function nested($a, $b) {
    if ($a)
        if ($b) {
            return 1;
        } else {
            echo 'x';
        }
    return 2;
}

if (function_exists('polyfill')) {
    return;
} else {
    function polyfill() {}
}
if (PHP_VERSION_ID > 80000) {
    return;
}
echo 'old';
