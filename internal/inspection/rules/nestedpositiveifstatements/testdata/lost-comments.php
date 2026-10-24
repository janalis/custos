<?php
function lost($a, $b, $c)
{
    // The fix could not keep these comments: none is offered.
    if ($a) {
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($b) {
            go();
        }
        // after the inner if
    }
    if ($a) {
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($b) {
            go();
        } else {
            other(); // in the discarded else
        }
    } else {
        other();
    }
    if ($a) {
        // before an unbraced body
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($b) go();
    }
    if ($a) {
        go();
    } else {
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($c) {
            go();
        }
        /* after it */
    }
}
