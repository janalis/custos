<?php

namespace Money {
    function intval($v, int $base = 10): int { return 0; }

    function convert($v, $s)
    {
        // Money\intval() is a user function.
        $a = intval($v);
        $b = (float)$v;
        $c = (int)$v;
        $v = (bool)$v;
        $d = (string)$s;
    }
}
