<?php
class Base
{
    public function __construct($id = 0) {}
    public function __clone() {}
    private function __destruct() {}
}

class Point extends Base
{
    public function __construct($id = 0) { parent::__construct($id); }
    static public function __clone() { parent::__clone(); }
    public function __destruct() {}
    public function __get() {}
    protected function __call($verb, $args) {}
    public function __callStatic($verb, $args) {}
    public function __fetchAll() {}
    public function __soapCall() {}
    public function _helper() {}
}

class Line extends Base
{
    public function __construct($id = 0) {}
    #[\Override]
    public function __clone() {}
    public function __toString()
    {
        $f = function () { return 1; };
        if (rand(0, 1)) {
            return $this->label;
        }
        return 42;
    }
    public function __wakeup()
    {
        $f = fn() => 1;
        return true;
    }
    public function __sleep() {}
    public function __debugInfo(): string { return ''; }
    public function __isset(&$key) {}
    public function __set($key, $value) {}
    public function __unset($key) {}
}

class Shape
{
    public static function __set_state($props) { return new static(); }
    public function __invoke() {}
    public function __autoload($cls) {}
    public function __serialize(): array { return []; }
    public function __unserialize() {}
}

interface Printable { public function __toString(); }

abstract class Partial { abstract public function __get($k); }
