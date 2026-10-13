<?php
class Base {
    public function render() {}
}

class Plain {
    public function h() {
        return static function () { return parent::render(); };
    }
}

class Card extends Base {
    public function h($m) {
        return [
            #[Pure] function () { return $this; },
            #[Pure] #[Other] fn () => $this,
            static function () use ($m) { return parent::$m(); },
        ];
    }
}
