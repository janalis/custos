<?php

namespace Lists {
    function array_values(array $a): array { return $a; }
    function range($a, $b): array { return []; }

    function check(array $rows): array
    {
        return [
            // Lists\array_values() and Lists\range() are user functions.
            array_values($rows) === $rows,
            \array_keys($rows) === range(0, \count($rows) - 1),
            <weak_warning descr="Replace with '\array_is_list($rows)'.">\Array_Values($rows) === $rows</weak_warning>,
            <weak_warning descr="Replace with '\array_is_list($rows)'.">\ARRAY_KEYS($rows) === \Range(0, COUNT($rows) - 1)</weak_warning>,
        ];
    }
}
