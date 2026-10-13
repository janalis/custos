<?php
namespace Text;

use function Lib\{strlen, helper};

function run($s, $fn) {
    $a = strlen($s);
    $b = <weak_warning descr="Write '\count(...)' to allow compile-time binding.">count($s)</weak_warning>;
    $c = $fn($s);
    $d = \array_filter($s, ...);
    return [$a, $b, $c, $d];
}
