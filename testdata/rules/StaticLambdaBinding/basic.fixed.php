<?php
class Base {
    public function render() {}
    public static function make() {}
}

class Widget extends Base {
    private $label = 'w';

    public function handlers() {
        return [
            function () { return $this->label; },
            function () { return strtoupper($this->label . $this->label); },
            fn ($s) => $s . $this->label,
            static function () { return fn () => 1; },
            static function () { return function () { return $this->label; }; },
            function ($x) { parent::render($x); },
            fn () => parent::render(),
            static function () { return parent::make(); },
        ];
    }
}
