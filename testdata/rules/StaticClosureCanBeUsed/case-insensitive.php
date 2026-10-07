<?php
abstract class Shape {
    public function area() {}
}

final class Square extends Shape {
    public function cases() {
        $inst = array_map(function ($n) { return PARENT::area(); }, [1]);
        $also = array_map(function ($n) { return Parent::Area(); }, [2]);

        $loose = <weak_warning descr="Closure does not use $this; declare it static.">function</weak_warning> () { return 5; };
        $moved = CLOSURE::BIND($loose, null, self::class);

        $free = <weak_warning descr="Closure does not use $this; declare it static.">function</weak_warning> () { return 6; };
        $free->BindTo(null);

        $kept = function () { return 8; };
        Closure::BIND($kept, $this);
    }
}
