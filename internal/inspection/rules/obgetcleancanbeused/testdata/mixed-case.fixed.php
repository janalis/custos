<?php
function capture(callable $fn) {
    ob_start();
    $fn();
    $out = ob_get_clean();
    return $out;
}
