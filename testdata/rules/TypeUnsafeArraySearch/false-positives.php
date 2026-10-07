<?php
function f(int $id, array $ids, string $name) {
    /** @var int[] $list */
    $list = load();
    return [
        in_array($id, $list),
        in_array($name, ['a', "b$name", ' c ']),
        in_array($id, $ids, true),
        array_search($name, ['x', 'y']),
    ];
}
