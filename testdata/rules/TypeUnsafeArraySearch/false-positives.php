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

// A spread passes an unknown number of arguments, the strict flag included.
function has($needle, ...$rest) { return in_array($needle, ...$rest); }
function spread(array $args, array $more) {
    return [
        in_array(...$args),
        in_array(...$args, ...$more),
        array_search(...$args, ...$more),
    ];
}
