<?php
function renderPanel($tpl) {
    ob_start();
    echo ob_get_contents();
    ob_end_clean();

    ob_start();
    $copy = ob_get_contents();
    flush_logs();
    ob_end_clean();

    ob_start();
    $last = ob_get_contents();
    $ok = ob_end_clean();

    ob_start();
    $doc = ob_get_contents();
    /** separator */
    ob_end_clean();

    ob_start();
    $q = Other\ob_get_contents();
    ob_end_clean();

    ob_start();
    $m = $buf->ob_get_contents();
    ob_end_clean();

    ob_start();
    $s = ob_get_contents();
    @ob_end_clean();

    return [$copy, $last, $ok, $doc, $q, $m, $s];
}
