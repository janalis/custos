<?php
class Other
{
    public $cache;
}

class Repo
{
    private $cache;
    private static $shared;

    public function __construct(Other $o)
    {
        $this->cache = [];
        self::$shared = 1;
        $o->cache = 2;
    }

    public function get()
    {
        return fn() => $this->cache;
    }
}

interface Shape
{
    public function __construct();
}

trait Holder
{
    private $x;
    public function __construct() { $this->x = 1; }
}

class NoCtor
{
    private $y;
    public function set() { $this->y = 1; }
}
