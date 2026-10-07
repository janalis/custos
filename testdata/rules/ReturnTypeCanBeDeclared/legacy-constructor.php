<?php
// Below PHP 8.0 a method named like its class is the constructor when the
// class has no __construct: declaring a return type on it is fatal.
class Kettle {
    private $litres;
    public function Kettle($litres) { $this->litres = $litres; }
    public function <weak_warning descr="Declare ': int' as the return type.">litres</weak_warning>() { return 2; }
}

class Teapot {
    public function __construct() {}
    // Not the constructor: __construct wins.
    public function <weak_warning descr="Declare ': void' as the return type.">teapot</weak_warning>() { echo 'brew'; }
}

interface Pourable {
    /** @return string */
    public function <weak_warning descr="Declare ': string' as the return type.">pourable</weak_warning>();
}
