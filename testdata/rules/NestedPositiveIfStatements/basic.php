<?php
function merge_cases($u, $v, $w) {
    if ($u) {
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($v) {
            go();
        }
    }

    if ($u && $v) {
        // explain
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($w > 2) {
            run();
        }
    }

    if ($u ?? $v) {
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($w) { run(); }
    }

    if ($u = load()) {
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($v) run();
    }

    if ($u) {
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($v) { run(); } else { stop(); }
    } else { stop(); }

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
    } else {
        // fallback
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($v) { stop(); } else {
            halt();
        }
    }
}
