<?php
trait HasColor {
    public $color = 'red';
    private $secret = 1;
}
trait Nested {
    use HasColor;
}

class Plain {
    public $color = 'red';
}

class Undecidable {
    use HasColor;
    public $color = self::RED;
    const RED = 'red';
}

class Base {
    private $secret = 2;
}

class Child extends Base {
    use HasColor;
    public $other;
}

interface Iface {}
