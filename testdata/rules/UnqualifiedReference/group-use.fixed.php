<?php
namespace Text;

use function Lib\{strlen, helper};

function run($s, $fn) {
    $a = strlen($s);
    $b = \count($s);
    $c = $fn($s);
    $d = \array_filter($s, ...);
    return [$a, $b, $c, $d];
}
