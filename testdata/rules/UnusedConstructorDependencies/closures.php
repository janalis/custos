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
