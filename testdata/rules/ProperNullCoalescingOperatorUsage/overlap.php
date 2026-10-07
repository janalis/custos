<?php
/** @param int|string|null $v */
function f($v, ?int $n, ?array $list, mixed $m, $fn) {
    return [
        $v ?? 1,
        <weak_warning descr="Operand types of '??' do not match ([int] vs [array]).">$n ?? []</weak_warning>,
        $list ?? [1, 2],
        $m ?? 1,
        <weak_warning descr="'$fn()' alone is equivalent; drop the '?? null' fallback.">$fn() ?? null</weak_warning>,
        <weak_warning descr="'\Foo::make()' alone is equivalent; drop the '?? null' fallback.">\Foo::make() ?? null</weak_warning>,
    ];
}
