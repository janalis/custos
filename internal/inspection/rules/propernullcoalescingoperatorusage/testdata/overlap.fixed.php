<?php
/** @param int|string|null $v */
function f($v, ?int $n, ?array $list, mixed $m, $fn) {
    return [
        $v ?? 1,
        $n ?? [],
        $list ?? [1, 2],
        $m ?? 1,
        $fn(),
        \Foo::make(),
    ];
}
