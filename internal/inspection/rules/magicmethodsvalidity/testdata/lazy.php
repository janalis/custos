<?php
final class LazyValue
{
    public readonly object $service;

    public function __construct(private \Closure $initializer)
    {
        unset($this->service, $other[0], $this->{'x'});
    }

    public function __get($name)
    {
        return $this->service = ($this->initializer)();
    }
}

final class Eager
{
    public function reset($other)
    {
        unset($other->service, $GLOBALS['x']);
    }

    public function <error descr="__get needs a companion __set method.">__get</error>($name)
    {
        return null;
    }
}
