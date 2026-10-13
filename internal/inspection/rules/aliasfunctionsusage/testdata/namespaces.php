<?php
namespace Shop\Util {
    function chop($s) { return $s; }

    $a = chop($label);
    $b = \<warning descr="Use 'rtrim(...)' instead of the alias 'chop(...)'.">chop</warning>($label);
}

namespace Shop\Web {
    use function Shop\Util\chop;

    $c = chop($label);
    $d = \<warning descr="Use 'rtrim(...)' instead of the alias 'chop(...)'.">chop</warning>($label);
    $e = <warning descr="Use 'current(...)' instead of the alias 'pos(...)'.">pos</warning>($stack);
}
