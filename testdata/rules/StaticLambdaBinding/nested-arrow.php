<?php
class Base {
    public function render() {}
}

class Panel extends Base {
    private $title = 't';

    public function handlers() {
        return [
            static function () { return fn () => <error descr="Static closures have no $this; remove 'static' or avoid $this.">$this</error>->title; },
            static function () { return fn () => fn () => <error descr="Calling an instance method of the parent class requires an object; this closure is static.">parent::render()</error>; },
            static function () { <error descr="Calling an instance method of the parent class requires an object; this closure is static.">parent::render()</error>; return <error descr="Static closures have no $this; remove 'static' or avoid $this.">$this</error>->title . <error descr="Static closures have no $this; remove 'static' or avoid $this.">$this</error>->title; },
            static function () { return static fn () => <error descr="Static closures have no $this; remove 'static' or avoid $this.">$this</error>->title; },
            function () { return fn () => $this->title; },
            static function () { return function () { return fn () => $this->title; }; },
        ];
    }
}
