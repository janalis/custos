<?php
class Base {
    public function render() {}
    public static function make() {}
    const C = 1;
}

class Widget extends Base {
    public function handlers() {
        return [
            static function () { return [parent::make(), parent::C, self::render(), parent::unknown()]; },
            static function () { return function () { parent::render(); }; },
            static function ($x) { return $x; },
        ];
    }
}
