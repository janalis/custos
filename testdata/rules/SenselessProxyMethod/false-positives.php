<?php
abstract class Base
{
    abstract public function run($a);
    public function swap($a, $b) {}
    public function other($a) {}
    final public function locked($a) {}
    public function __construct(protected $x = 1) {}
    public function attr($a) {}
    public function twice($a) {}
}

class Child extends Base
{
    public function run($a) { parent::run($a); }
    public function swap($a, $b) { parent::swap($b, $a); }
    public function other($a) { parent::swap($a, 1); }
    protected function attr(#[Sensitive] $a) { parent::attr($a); }
    public function __construct(private $x = 1) { parent::__construct($x); }
    public function twice($a) { parent::twice($a); return 1; }
    public function more($a) { return parent::more($a); }
    private function hidden($a) { parent::hidden($a); }
    public function unpack(...$a) { parent::unpack(...$a); }
}

trait T { public function swap($a, $b) { parent::swap($a, $b); } }
class Orphan extends Missing { public function m($a) { parent::m($a); } }
