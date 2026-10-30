<?php
function parents($strict, $u, $v, $w)
{
    if (!$row = load()) {
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($strict) {
            report();
        }
    }

    if (@$u = $v) {
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($w) { run(); }
    }

    if ((bool) $u = $v) {
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($w) { run(); }
    }

    if (print "p") {
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($w) { run(); }
    }

    if ($u && $row = load()) {
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($w) { run(); }
    }

    if (!$u) {
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($w) { run(); }
    }

    if ($u instanceof Message) {
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($w) { run(); }
    }
}

function children($u, $a, $b, $c)
{
    if ($u) {
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($a ? $b : $c) {
            run();
        }
    }

    if ($u) {
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($a ?? $b) { run(); }
    }

    if ($u) {
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($a xor $b) { run(); }
    }

    if ($u) {
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($a and $b) { run(); }
    }

    if ($u) {
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($a = $b) { run(); }
    }
}
