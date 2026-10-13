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
            static function () { return strtoupper(<error descr="Static closures have no $this; remove 'static' or avoid $this.">$this</error>->label . <error descr="Static closures have no $this; remove 'static' or avoid $this.">$this</error>->label); },
            static fn ($s) => $s . <error descr="Static closures have no $this; remove 'static' or avoid $this.">$this</error>->label,
            static function () { return fn () => 1; },
            static function () { return function () { return $this->label; }; },
            static function ($x) { <error descr="Calling an instance method of the parent class requires an object; this closure is static.">parent::render($x)</error>; },
            static fn () => <error descr="Calling an instance method of the parent class requires an object; this closure is static.">parent::render()</error>,
            static function () { return parent::make(); },
        ];
    }
}
