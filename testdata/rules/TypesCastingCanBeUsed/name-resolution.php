<?php

namespace Money {
    function intval($v, int $base = 10): int { return 0; }

    function convert($v, $s)
    {
        // Money\intval() is a user function.
        $a = intval($v);
        $b = <weak_warning descr="Use '(float)$v' instead (a cast is clearer and faster).">FloatVal($v)</weak_warning>;
        $c = <weak_warning descr="Use '(int)$v' instead (a cast is clearer and faster).">\INTVAL($v)</weak_warning>;
        <weak_warning descr="Use '$v = (bool)$v' instead (a cast is clearer and faster).">SetType($v, 'bool')</weak_warning>;
        $d = <weak_warning descr="Use '(string)$s' instead of calling __toString() directly.">$s->__TOSTRING()</weak_warning>;
    }
}
