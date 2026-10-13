<?php
// Reported, but not rewritten: code of the same statement still runs after
// the read and would then see a closed buffer.
function later($o, $list, $acc) {
    ob_start();
    print <warning descr="Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().">ob_get_contents()</warning>;
    ob_end_clean();

    ob_start();
    $r = <warning descr="Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().">ob_get_contents()</warning> . ob_get_level();
    ob_end_clean();

    ob_start();
    $r = <warning descr="Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().">ob_get_contents()</warning> || ob_get_level();
    ob_end_clean();

    ob_start();
    record(<warning descr="Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().">ob_get_contents()</warning>);
    ob_end_clean();

    ob_start();
    $o->html = <warning descr="Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().">ob_get_contents()</warning>;
    ob_end_clean();

    ob_start();
    $list[] = <warning descr="Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().">ob_get_contents()</warning>;
    ob_end_clean();

    ob_start();
    $$acc = <warning descr="Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().">ob_get_contents()</warning>;
    ob_end_clean();

    ob_start();
    $acc .= <warning descr="Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().">ob_get_contents()</warning>;
    ob_end_clean();

    ob_start();
    $n = -<warning descr="Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().">ob_get_contents()</warning>;
    ob_end_clean();

    ob_start();
    $n = <warning descr="Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().">ob_get_contents()</warning> ? 1 : notify();
    ob_end_clean();

    ob_start();
    $n = trim(<warning descr="Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().">ob_get_contents()</warning>, $acc);
    ob_end_clean();

    ob_start();
    $n = strlen(...<warning descr="Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().">ob_get_contents()</warning>);
    ob_end_clean();

    ob_start();
    $n = (<warning descr="Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().">ob_get_contents()</warning>)();
    ob_end_clean();

    ob_start();
    $n = $o->parts[<warning descr="Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().">ob_get_contents()</warning>]?->name;
    ob_end_clean();

    ob_start();
    $n = wrap(<warning descr="Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().">ob_get_contents()</warning>)?->body->render();
    ob_end_clean();

    return [$r, $n];
}

// Rewritten: only inert wrappers and a plain local assignment follow.
function inert() {
    ob_start();
    $a = $b = (string) (<warning descr="Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().">ob_get_contents()</warning>);
    ob_end_clean();

    ob_start();
    $c = RTrim(trim(<warning descr="Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().">ob_get_contents()</warning>));
    ob_end_clean();

    ob_start();
    <warning descr="Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().">ob_get_contents()</warning>;
    ob_end_clean();

    return [$a, $b, $c];
}
