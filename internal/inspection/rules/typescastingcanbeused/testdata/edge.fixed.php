<?php
namespace {
function edges($v, $type, $o) {
    settype($v);
    settype($v, $type);
    settype($v, 'resource');
    $a = intval($v, 10, 1);
    $b = strval(...$o);
    $c = ((string)$o)->length;
    $d = ((string)$o)->trim();
    $e = ((int)$v) ** 2;
    $f = 2 ** ((int)$v);
    $g = $fn($v);
    $h = $o->format();
}
}

namespace Conditional {

if (!\function_exists('Conditional\settype')) {
    function settype(&$v, $t) {}
}

function local($v) {
    settype($v, 'int'); // the conditional Conditional\settype() wins
}
}
