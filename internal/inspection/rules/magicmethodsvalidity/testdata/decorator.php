<?php

namespace App;

class Event
{
    private $name;

    public function __construct($name)
    {
        $this->name = $name;
    }
}

// A decorator wraps an instance of its parent instead of initialising it.
final class LazyEvent extends Event
{
    public function __construct(private parent $event) {}
}

final class TracedEvent extends Event
{
    public function __construct(private ?Event $event, $tag) {}
}

final class Renamed extends Event
{
    public function <error descr="__construct does not call Event::__construct().">__construct</error>(array $names, \Closure|Event $inner, $name) {}
}
