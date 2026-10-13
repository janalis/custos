<?php
class Base {
    public function render() {}
}

class Panel extends Base {
    private $title = 't';

    public function handlers() {
        return [
            function () { return fn () => $this->title; },
            function () { return fn () => fn () => parent::render(); },
            function () { parent::render(); return $this->title . $this->title; },
            static function () { return fn () => $this->title; },
            function () { return fn () => $this->title; },
            static function () { return function () { return fn () => $this->title; }; },
        ];
    }
}
