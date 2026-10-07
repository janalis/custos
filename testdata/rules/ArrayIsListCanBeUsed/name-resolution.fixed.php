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
            \array_is_list($rows),
            $rows !== [] && \array_is_list($rows),
        ];
    }
}
