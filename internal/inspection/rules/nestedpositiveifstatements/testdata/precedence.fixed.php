<?php
function parents($strict, $u, $v, $w)
{
    if ((!$row = load()) && $strict) {
            report();
        }

    if ((@$u = $v) && $w) { run(); }

    if (((bool) $u = $v) && $w) { run(); }

    if ((print "p") && $w) { run(); }

    if (($u && $row = load()) && $w) { run(); }

    if (!$u && $w) { run(); }

    if ($u instanceof Message && $w) { run(); }
}

function children($u, $a, $b, $c)
{
    if ($u && ($a ? $b : $c)) {
            run();
        }

    if ($u && ($a ?? $b)) { run(); }

    if ($u && ($a xor $b)) { run(); }

    if ($u && $a and $b) { run(); }

    if ($u && $a = $b) { run(); }
}
