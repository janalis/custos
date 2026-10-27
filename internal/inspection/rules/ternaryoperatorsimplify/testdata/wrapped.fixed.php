<?php
function f($v, $w) {
    return [
        'a' => (!empty($v) && $v == TRUE),
        'b' => (!($v || $w)),
        'c' => ($v || $w),
    ];
}
