<?php

namespace Pipeline {
    function array_map($fn, array $items) { return $items; }

    function run(array $items)
    {
        // Pipeline\array_map() does not invoke its first argument.
        $own = array_map('compact', $items);
        $global = \Array_Map(<warning descr="'compact' reads the caller scope and cannot be invoked indirectly since PHP 7.1.">'compact'</warning>, $items);
    }
}
