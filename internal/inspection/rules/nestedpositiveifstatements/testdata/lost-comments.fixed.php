<?php
function lost($a, $b, $c)
{
    // The fix could not keep these comments: none is offered.
    if ($a) {
        if ($b) {
            go();
        }
        // after the inner if
    }
    if ($a) {
        if ($b) {
            go();
        } else {
            other(); // in the discarded else
        }
    } else {
        other();
    }
    if ($a) {
        // before an unbraced body
        if ($b) go();
    }
    if ($a) {
        go();
    } else {
        if ($c) {
            go();
        }
        /* after it */
    }
}
