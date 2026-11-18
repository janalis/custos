<?php
namespace Shop\Util {
    function chop($s) { return $s; }

    $a = chop($label);
    $b = \rtrim($label);
}

namespace Shop\Web {
    use function Shop\Util\chop;

    $c = chop($label);
    $d = \rtrim($label);
    $e = current($stack);
}
