<?php
function f($a) {
    if ($a) {
        return 1;
    }
    /* keep me */
    work();
    if ($a > 1) {
        return 2;
    }
    // note
    if ($a > 0) {
        work();
    }
}
function g($a) { if ($a) { return 1; }
work(); // done
 work(); return 2; }
