<?php
namespace Text;

use function strlen as len;
use function Lib\measure as count;
use const PHP_EOL as EOL;

function run($s, $list) {
    $a = <weak_warning descr="Write '\strlen(...)' to allow compile-time binding.">strlen($s)</weak_warning>;
    $b = len($s);
    $c = count($list);
    $d = <weak_warning descr="Write '\PHP_EOL' to allow compile-time binding.">PHP_EOL</weak_warning>;
    return [$a, $b, $c, $d];
}
