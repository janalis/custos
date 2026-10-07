<?php
trait Writes
{
    public function __set($k, $v) {}
    public function __get($k) { return 1; }
    public function __isset($k) { return true; }
}

class Plain
{
    use Writes;
    public function regular() {}
    public function __get($k) { return 1; }
    public function __invoke() {}
    public function <error descr="__toString must not declare parameters.">__toString</error>($x)
    {
        if ($x) {
            <error descr="__toString must return string; got ''.">return;</error>
        }
        return 'a';
    }
    public static function __set_state($a): self { return new self(); }
    /** @var string[] */
    private $fields = [];
    public function __sleep() { return $this->fields; }
}

class Orphan extends MissingBase
{
    public function __construct() {}
}

class Base
{
    public function __construct() {}
}

class Child extends Base
{
    public function __construct(Base $other) { $other->__construct(); }
}

class Statics
{
    public static function <error descr="__invoke must not be static.">__invoke</error>() {}
}
