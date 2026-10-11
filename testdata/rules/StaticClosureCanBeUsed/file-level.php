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

$called = <weak_warning descr="Closure does not use $this; declare it static.">function</weak_warning> () { return 2; };
echo $called();

$detached = <weak_warning descr="Closure does not use $this; declare it static.">function</weak_warning> () { return 3; };
$detached->bindTo(null);

$lazy = <weak_warning descr="Closure does not use $this; declare it static.">fn</weak_warning> () => $inner = function () { return 4; };
