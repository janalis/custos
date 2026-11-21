<?php
function f($price, $o, $x) {
    // Before 7.0 a parenthesized cast cannot be dereferenced.
    $first = strval($price)[0];
    $len = strval($o)->length;
    $text = (string)$price;
    $square = ((int)$x) ** 2;
    return array($first, $len, $text, $square);
}
