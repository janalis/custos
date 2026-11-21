<?php
function f($price, $o, $x) {
    // Before 7.0 a parenthesized cast cannot be dereferenced.
    $first = strval($price)[0];
    $len = strval($o)->length;
    $text = <weak_warning descr="Use '(string)$price' instead (a cast is clearer and faster).">strval($price)</weak_warning>;
    $square = <weak_warning descr="Use '((int)$x)' instead (a cast is clearer and faster).">intval($x)</weak_warning> ** 2;
    return array($first, $len, $text, $square);
}
