<?php

namespace Shop\Hooks;

use Acme\Absent\Listener;

final class Dispatcher
{
    public function __invoke(): void {}
}

final class Plain {}

class Factory
{
    public function missing(?callable $h = null)
    {
        return $h ?? new Listener();
    }

    public function invokable(?callable $h = null)
    {
        return $h ?? new Dispatcher();
    }

    public function invokableLeft(?Dispatcher $d = null, ?callable $h = null)
    {
        return $d ?? $h;
    }

    public function plain(?callable $h = null)
    {
        return <weak_warning descr="Operand types of '??' do not match ([callable] vs [\Shop\Hooks\Plain]).">$h ?? new Plain()</weak_warning>;
    }
}
