<?php
function inspect(array $cfg) {
    return [
        array_values($cfg) == $cfg,
        $cfg != array_values($cfg),
        array_keys($cfg) == range(0, count($cfg) - 1),
        range(0, count($cfg) - 1) != array_keys($cfg),
        <weak_warning descr="Replace with 'array_is_list($cfg)'.">array_values($cfg) === $cfg</weak_warning>,
    ];
}
