<?php
abstract class Shape {
    public function area() {}
}

final class Square extends Shape {
    public function cases() {
        $inst = array_map(function ($n) { return PARENT::area(); }, [1]);
        $also = array_map(function ($n) { return Parent::Area(); }, [2]);

        $loose = static function () { return 5; };
        $moved = CLOSURE::BIND($loose, null, self::class);

        $free = static function () { return 6; };
        $free->BindTo(null);

        $kept = function () { return 8; };
        Closure::BIND($kept, $this);
    }
}
