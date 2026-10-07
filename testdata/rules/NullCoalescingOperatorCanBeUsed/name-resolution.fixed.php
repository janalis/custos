<?php

namespace Cfg {
    function array_key_exists($k, $a): bool { return true; }

    function read(array $o)
    {
        // Cfg\array_key_exists() is a user function.
        $a = array_key_exists('k', $o) ? $o['k'] : null;
        $b = $o['k'] ?? \NULL;
        $c = $o ?? 1;
    }
}
