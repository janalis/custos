<?php
function f($s) {
    $s = str_replace(array('a', 'c', 'd'), array('b', 'e', 'e'), $s);
    return $s;
}
