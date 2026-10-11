<?php
function colon_children($u, $v)
{
    // The colon-syntax body has no closing keyword of its own to carry over.
    if ($u) {
        if ($v): run(); endif;
    }

    if ($u) {
        if ($v): run(); else: stop(); endif;
    } else {
        stop();
    }

    if ($u):
        if ($v) {
            run();
        }
    endif;
}
