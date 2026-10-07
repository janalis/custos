<?php
class Mode { const FAST = 1; }
class Base {
    protected $mode = Mode::FAST;
}
class Child extends Base {
    protected $mode = <weak_warning descr="Default repeats the inherited value; remove it.">mode::FAST</weak_warning>;
}
class Own {
    private $mode = Mode::FAST;
    public function __construct() {
        <weak_warning descr="Assignment writes the property's default value; remove it.">$this->mode = MODE::FAST;</weak_warning>
    }
}
