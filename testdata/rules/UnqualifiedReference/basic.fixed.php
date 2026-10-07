<?php
namespace App;

function run($s, $list) {
    $a = \strlen($s);
    $b = trim($s);
    $c = PHP_EOL;
    $d = \call_user_func('\is_int', $s);
    $e = \array_filter($list, 'trim');
    return [$a, $b, $c, $d, $e];
}
