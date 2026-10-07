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
