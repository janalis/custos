<?php
class Registry {
    private $items = [];

    public function handlers() {
        return [
            static function () { return isset($this) ? $this->items : null; },
            function () { return isset($this->items) || empty($this['k']); },
            fn () => $this ?? null,
            static function () { $f = function () { return empty($this); }; return $this->items; },
        ];
    }
}
