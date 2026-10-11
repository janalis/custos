<?php
function matched($x) {
    return (match ($x) { default => new Box })->size;
}
function otherNames() {
    [$x, $y] = explode(':', 'a:b');
    return compact('x', 'y');
}
function required() {
    return require 'config.php';
}
class Registry {
    public function &slots() {
        $read = function () {
            return load();
        };
        return $this->slots;
    }
}
