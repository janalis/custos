<?php
namespace App;

function run($s, $list) {
    $a = <weak_warning descr="Write '\strlen(...)' to allow compile-time binding.">strlen($s)</weak_warning>;
    $b = trim($s);
    $c = PHP_EOL;
    $d = <weak_warning descr="Write '\call_user_func(...)' to allow compile-time binding.">call_user_func(<weak_warning descr="Write '\is_int' to allow compile-time binding.">'is_int'</weak_warning>, $s)</weak_warning>;
    $e = \array_filter($list, 'trim');
    return [$a, $b, $c, $d, $e];
}
