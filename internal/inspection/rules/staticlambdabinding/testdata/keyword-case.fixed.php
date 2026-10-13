<?php
class Shape {
    public function draw() {}
    public static function unit() {}
}

class Circle extends Shape {
    public function painters() {
        return [
            function () { PARENT::draw(); },
            static function () { return Parent::unit(); },
        ];
    }
}
