<?php
function merge_cases($u, $v, $w) {
    if ($u && $v) {
            go();
        }

    if ($u && $v && $w > 2) {
        // explain
            run();
        }

    if (($u ?? $v) && $w) { run(); }

    if (($u = load()) && $v) run();

    if ($u && $v) { run(); } else { stop(); }

    if ($u) {
        if ($v) {}
    } else {
        stop();
    }

    if ($u || $v) {
        if ($w) {}
    }

    if ($u) {
        if ($v) {}
        else { stop(); }
    }

    if ($u) {
        if ($v) {}
        tick();
    }

    if ($u) {
        if ($v) {}
    } elseif ($w) {
    }

    if ($u) {
    } elseif ($w) {
        if ($v) {}
    }
}

function else_cases($u, $v) {
    if ($u) {
        run();
    } else if ($v) {
        // fallback
        stop(); } else {
            halt();
        }
}
