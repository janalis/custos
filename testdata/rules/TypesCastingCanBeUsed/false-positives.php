<?php
namespace App {
    function settype(&$v, $t) {}

    settype($x, 'int');
    \Other\settype($x, 'int');
    $a = doubleval($x);
    $b = intval(...$xs);
    $c = intval($x, $base);
    $d = " $x";
    $e = "$x ";
    $f = 'x';
    $g = <<<TXT
$x
TXT;
    $h = "${x}";
    $i = $o?->__toString();
    $j = Foo::__toString();
    settype($x, 'INT');
    settype($x, $type);
}
