<?php
class Shape {
    public function draw() {}
    public static function unit() {}
}

class Circle extends Shape {
    public function painters() {
        return [
            static function () { <error descr="Calling an instance method of the parent class requires an object; this closure is static.">PARENT::draw()</error>; },
            static function () { return Parent::unit(); },
        ];
    }
}
