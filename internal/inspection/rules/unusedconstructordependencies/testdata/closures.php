<?php
class Dispatcher
{
    private $handler;
    private $hook;
    private $counter;
    private $plain;
    public $onError;

    public function __construct($handler, $plain)
    {
        $this->handler = $handler;
        $this->onError = fn($e) => $this->handler->handle($e);
        <weak_warning descr="Private property is only used in the constructor; likely dead code.">$this->hook</weak_warning> = false;
        $register = function () { $this->hook = true; };
        $this->counter = 0;
        $tick = function () { $this->counter++; };
        <weak_warning descr="Private property is only used in the constructor; likely dead code.">$this->plain</weak_warning> = $plain;
    }
}

// A closure created in the constructor may read any property later.
class Resolver
{
    private $a;
    public $getter;

    public function __construct($a)
    {
        $this->a = $a;
        $this->getter = fn($n) => $this->$n;
    }
}

class Exporter
{
    private $a;
    public $dump;

    public function __construct($a)
    {
        $this->a = $a;
        $this->dump = function () { return get_object_vars($this); };
    }
}

// The constructor's own scope is not a later use.
class Snapshot
{
    private $name;
    public $state;

    public function __construct($name)
    {
        <weak_warning descr="Private property is only used in the constructor; likely dead code.">$this->name</weak_warning> = $name;
        $this->state = get_object_vars($this);
    }
}
