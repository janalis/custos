<?php
namespace App {
    function substr($s, $a, $l = null) { return ''; }
    function strlen($s) { return 0; }

    $a = substr($name, 0, \strlen($name) - 2);
    $b = \substr($name, 0, strlen($name) - 2);
}

namespace {
    $c = SubStr($name, 0, <warning descr="Pass '-2' as the length instead.">STRLEN($name) - 2</warning>);
}
