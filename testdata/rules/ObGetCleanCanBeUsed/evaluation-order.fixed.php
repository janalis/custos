<?php
// Reported, but not rewritten: code of the same statement still runs after
// the read and would then see a closed buffer.
function later($o, $list, $acc) {
    ob_start();
    print ob_get_contents();
    ob_end_clean();

    ob_start();
    $r = ob_get_contents() . ob_get_level();
    ob_end_clean();

    ob_start();
    $r = ob_get_contents() || ob_get_level();
    ob_end_clean();

    ob_start();
    record(ob_get_contents());
    ob_end_clean();

    ob_start();
    $o->html = ob_get_contents();
    ob_end_clean();

    ob_start();
    $list[] = ob_get_contents();
    ob_end_clean();

    ob_start();
    $$acc = ob_get_contents();
    ob_end_clean();

    ob_start();
    $acc .= ob_get_contents();
    ob_end_clean();

    ob_start();
    $n = -ob_get_contents();
    ob_end_clean();

    ob_start();
    $n = ob_get_contents() ? 1 : notify();
    ob_end_clean();

    ob_start();
    $n = trim(ob_get_contents(), $acc);
    ob_end_clean();

    ob_start();
    $n = strlen(...ob_get_contents());
    ob_end_clean();

    ob_start();
    $n = (ob_get_contents())();
    ob_end_clean();

    ob_start();
    $n = $o->parts[ob_get_contents()]?->name;
    ob_end_clean();

    ob_start();
    $n = wrap(ob_get_contents())?->body->render();
    ob_end_clean();

    return [$r, $n];
}

// Rewritten: only inert wrappers and a plain local assignment follow.
function inert() {
    ob_start();
    $a = $b = (string) (ob_get_clean());

    ob_start();
    $c = RTrim(trim(ob_get_clean()));

    ob_start();
    ob_get_clean();

    return [$a, $b, $c];
}
