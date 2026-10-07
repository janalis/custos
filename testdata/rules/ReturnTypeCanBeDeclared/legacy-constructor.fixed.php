<?php
// Below PHP 8.0 a method named like its class is the constructor when the
// class has no __construct: declaring a return type on it is fatal.
class Kettle {
    private $litres;
    public function Kettle($litres) { $this->litres = $litres; }
    public function litres(): int { return 2; }
}

class Teapot {
    public function __construct() {}
    // Not the constructor: __construct wins.
    public function teapot(): void { echo 'brew'; }
}

interface Pourable {
    /** @return string */
    public function pourable(): string;
}
