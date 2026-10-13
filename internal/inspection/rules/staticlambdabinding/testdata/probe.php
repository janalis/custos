<?php
class Registry {
    private $items = [];

    public function handlers() {
        return [
            static function () { return isset($this) ? <error descr="Static closures have no $this; remove 'static' or avoid $this.">$this</error>->items : null; },
            static function () { return isset(<error descr="Static closures have no $this; remove 'static' or avoid $this.">$this</error>->items) || empty(<error descr="Static closures have no $this; remove 'static' or avoid $this.">$this</error>['k']); },
            static fn () => <error descr="Static closures have no $this; remove 'static' or avoid $this.">$this</error> ?? null,
            static function () { $f = function () { return empty($this); }; return <error descr="Static closures have no $this; remove 'static' or avoid $this.">$this</error>->items; },
        ];
    }
}
