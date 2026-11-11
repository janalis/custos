<?php
// The first element of the callable pair may hold a class name: a direct
// `$x->m()` call would then fail. Unknown, mixed, string and callable
// (which may be a string) receivers are not reported.
/**
 * @param mixed $mixed
 */
function receivers($unknown, $mixed, string $name, callable $cb, object $obj) {
    call_user_func([$unknown, 'run']);
    call_user_func([$mixed, 'run']);
    call_user_func([$name, 'run']);
    call_user_func([$cb, 'run']);
    /** @var array|callable $either */
    $either = source();
    call_user_func([$either, 'run']);
    <weak_warning descr="Call it directly: '$obj->run()'.">call_user_func([$obj, 'run'])</weak_warning>;
}
