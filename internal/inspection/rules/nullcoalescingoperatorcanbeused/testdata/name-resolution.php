<?php

namespace Cfg {
    function array_key_exists($k, $a): bool { return true; }

    function read(array $o)
    {
        // Cfg\array_key_exists() is a user function.
        $a = array_key_exists('k', $o) ? $o['k'] : null;
        $b = <weak_warning descr="Simplify to '$o['k'] ?? \NULL' using the null coalescing operator.">\Array_Key_Exists('k', $o) ? $o['k'] : \NULL</weak_warning>;
        $c = <weak_warning descr="Simplify to '$o ?? 1' using the null coalescing operator.">$o !== \null ? $o : 1</weak_warning>;
    }
}
