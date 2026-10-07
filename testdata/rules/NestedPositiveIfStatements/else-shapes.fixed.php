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

    if ($a && $b) {
            run();
        } else {
        // same statements, other comments
        stop();
    }
}
