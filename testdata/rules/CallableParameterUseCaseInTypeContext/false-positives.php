<?php
/**
 * @param mixed  $any
 * @param number $qty
 */
function loose($any, $qty, string $code, object $o, $plain) {
    return [is_int($any), is_float($qty), is_numeric($code), is_long($code), is_object($o), is_int($plain)];
}

function untyped($flag = null) {
    $flag = 'on';
}

class Node {}
function nodes(?Node $n, string $s, int $i) {
    $n = $n ?: new Node();
    $s = strstr($s, 'x');
    $i = abs($i);
    $i = microtime();
    return;
    $s = 1;
}

class FooTest {
    public function check(int $x) { return is_string($x); }
}

function normalizeEol(string $text, array $lines) {
    $text = str_replace(["\r\n", "\r"], replace: "\n", subject: $text);
    $text = preg_replace(subject: $text, pattern: '/\s+/', replacement: ' ');
    $lines = str_replace("\t", ' ', subject: $lines);
    return [$text, $lines];
}

function expiry(int $ttl, $extra) {
    // An operand of unknown type: the sum is not known to be a float.
    $ttl = time() + $extra;
    return $ttl;
}

function toUtf8(string $text) {
    // A string input converts to a string (or false), never an array.
    $text = mb_convert_encoding($text, 'UTF-8');
    return $text;
}
