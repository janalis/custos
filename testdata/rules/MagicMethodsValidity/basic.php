<?php
class Base
{
    public function __construct($id = 0) { $this->id = $id; }
    public function __clone() {}
    private function __destruct() {}
}

class Point extends Base
{
    public function __construct($id = 0) { parent::__construct($id); }
    static public function <error descr="__clone must not be static.">__clone</error>() { parent::__clone(); }
    public function __destruct() {}
    public function <error descr="__get must declare exactly 1 parameter(s)."><error descr="__get needs a companion __set method.">__get</error></error>() {}
    protected function <error descr="__call must be declared public.">__call</error>($verb, $args) {}
    public function <error descr="__callStatic must be declared static.">__callStatic</error>($verb, $args) {}
    public function <error descr="The '__' prefix is reserved for magic methods.">__fetchAll</error>() {}
    public function __soapCall() {}
    public function _helper() {}
}

class Line extends Base
{
    public function <error descr="__construct does not call Base::__construct().">__construct</error>($id = 0) {}
    #[\Override]
    public function __clone() {}
    public function __toString()
    {
        $f = function () { return 1; };
        if (rand(0, 1)) {
            return $this->label;
        }
        <error descr="__toString must return string; got 'int'.">return 42;</error>
    }
    public function __wakeup()
    {
        $f = fn() => 1;
        <error descr="__wakeup must not return a value.">return true;</error>
    }
    public function <error descr="__sleep must return array; got ''.">__sleep</error>() {}
    public function <error descr="__debugInfo must return array|null; got 'string'.">__debugInfo</error>(): string { return ''; }
    public function <error descr="__isset must not take parameters by reference.">__isset</error>(&$key) {}
    public function <error descr="__set needs a companion __get method.">__set</error>($key, $value) {}
    public function __unset($key) {}
}

class Shape
{
    public static function __set_state($props) { return new static(); }
    public function <error descr="'_invoke' is not magic; did you mean '__invoke'?">_invoke</error>() {}
    public function <error descr="__autoload is deprecated since PHP 7.2; use spl_autoload_register().">__autoload</error>($cls) {}
    public function __serialize(): array { return []; }
    public function <error descr="__unserialize must declare exactly 1 parameter(s).">__unserialize</error>() {}
}

interface Printable { public function __toString(); }

abstract class Partial { abstract public function __get($k); }
