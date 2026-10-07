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
            #[Pure] static function () { return <error descr="Static closures have no $this; remove 'static' or avoid $this.">$this</error>; },
            #[Pure] #[Other] static fn () => <error descr="Static closures have no $this; remove 'static' or avoid $this.">$this</error>,
            static function () use ($m) { return parent::$m(); },
        ];
    }
}
