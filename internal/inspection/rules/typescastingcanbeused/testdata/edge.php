<?php
namespace {
function edges($v, $type, $o) {
    settype($v);
    settype($v, $type);
    settype($v, 'resource');
    $a = intval($v, 10, 1);
    $b = strval(...$o);
    $c = <weak_warning descr="Use '((string)$o)' instead (a cast is clearer and faster).">strval($o)</weak_warning>->length;
    $d = <weak_warning descr="Use '((string)$o)' instead (a cast is clearer and faster).">strval($o)</weak_warning>->trim();
    $e = <weak_warning descr="Use '((int)$v)' instead (a cast is clearer and faster).">intval($v)</weak_warning> ** 2;
    $f = 2 ** <weak_warning descr="Use '((int)$v)' instead (a cast is clearer and faster).">intval($v)</weak_warning>;
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
