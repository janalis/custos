<?php
// Byte and character lengths are not interchangeable: no report.
function units($label, $k) {
    $a = mb_substr($label, 0, strlen($label) - 2);
    $b = substr($label, 0, mb_strlen($label) - 3);
    $c = substr($label, $k, mb_strlen($label) - $k);
    $d = mb_substr($label, 1, strlen($label) - 1, 'UTF-8');
    return [$a, $b, $c, $d];
}
