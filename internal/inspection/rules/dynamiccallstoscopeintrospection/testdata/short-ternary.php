<?php
function shortCalls($fallback, $rows) {
    $fn = 'extract' ?: $fallback;
    <warning descr="'extract' reads the caller scope and cannot be invoked indirectly since PHP 7.1.">$fn</warning>($rows);
    $ambiguous = 'extract' ?: 'trim';
    $ambiguous($rows);
}
