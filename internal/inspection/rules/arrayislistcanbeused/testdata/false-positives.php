<?php
function inspect(array $cfg, array $other) {
    return [
        array_values($cfg) === $other,
        array_values($cfg) === [],
        array_values($cfg) <> $cfg,
        array_values($cfg) < $cfg,
        array_keys($cfg) === range(1, count($cfg)),
        array_keys($cfg) === range(0, count($other) - 1),
        array_keys($cfg) === range(0, count($cfg) - 2),
        array_keys($cfg) === range(0, count($cfg) - 1, 1),
        array_keys($cfg) === range(0, count($cfg, 1) - 1),
        array_keys($cfg, 'on') === range(0, count($cfg) - 1),
        array_keys($cfg) === $cfg,
        array_keys($cfg) === range(0, count($cfg) + 1),
        array_keys($cfg) === range(0, $other - 1),
        (array_values($cfg)) === $cfg,
        array_values($cfg),
    ];
}
