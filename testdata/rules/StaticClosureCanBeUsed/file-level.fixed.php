<?php
class Ctx { private $secret = 's3'; }

$reader = function ($o) { return $o->secret; };
$bound = Closure::bind($reader, new Ctx, Ctx::class);

$binder = function ($o) { return $o->secret; };
$binder->bindTo(new Ctx, Ctx::class);

$shared = function () { return 1; };
function uses_shared() {
    global $shared;
    return $shared;
}

$called = static function () { return 2; };
echo $called();

$detached = static function () { return 3; };
$detached->bindTo(null);

$lazy = static fn () => $inner = function () { return 4; };
