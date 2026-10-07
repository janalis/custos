<?php
function inspect(array $cfg) {
    return [
        array_values($cfg) == $cfg,
        $cfg != array_values($cfg),
        array_keys($cfg) == range(0, count($cfg) - 1),
        range(0, count($cfg) - 1) != array_keys($cfg),
        array_is_list($cfg),
    ];
}
