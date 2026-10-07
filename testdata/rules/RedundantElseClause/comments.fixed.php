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
