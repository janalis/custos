<?php
function stamps($spec) {
    $a = time();
    $b = STRTOTIME($spec);
    return [$a, $b];
}
