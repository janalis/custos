<?php
function shapes($a, $b, $c)
{
    if ($a) {
        if ($b) {
            run();
        } else if ($c) {
            stop();
        }
    } else if ($c) {
        stop();
    }

    if ($a) {
        if ($b) {
            run();
        } else {
            stop();
        }
    } else {
        halt();
    }

    if ($a) {
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($b) {
            run();
        } else {
            stop();
        }
    } else {
        // same statements, other comments
        stop();
    }
}
