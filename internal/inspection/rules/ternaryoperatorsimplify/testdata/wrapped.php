<?php
function f($v, $w) {
    return [
        'a' => (<weak_warning descr="Replace the ternary with '!empty($v) && $v == TRUE'.">(!empty($v) && $v == TRUE) ? TRUE : FALSE</weak_warning>),
        'b' => (<weak_warning descr="Replace the ternary with '!($v || $w)'.">$v || $w ? false : true</weak_warning>),
        'c' => <weak_warning descr="Replace the ternary with '($v || $w)'.">$v || $w ? true : false</weak_warning>,
    ];
}
