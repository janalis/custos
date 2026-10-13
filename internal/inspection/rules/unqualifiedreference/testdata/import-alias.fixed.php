<?php
namespace Text;

use function strlen as len;
use function Lib\measure as count;
use const PHP_EOL as EOL;

function run($s, $list) {
    $a = \strlen($s);
    $b = len($s);
    $c = count($list);
    $d = \PHP_EOL;
    return [$a, $b, $c, $d];
}
