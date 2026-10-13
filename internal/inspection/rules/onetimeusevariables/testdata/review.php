<?php
function matched($x) {
    <warning descr="Variable $o is used only once; inline its value.">$o</warning> = match ($x) { default => new Box };
    return $o->size;
}
function otherNames() {
    <warning descr="Variable $parts is used only once; inline its value.">$parts</warning> = explode(':', 'a:b');
    [$x, $y] = $parts;
    return compact('x', 'y');
}
function required() {
    <warning descr="Variable $config is used only once; inline its value.">$config</warning> = require 'config.php';
    return $config;
}
class Registry {
    public function &slots() {
        $read = function () {
            <warning descr="Variable $copy is used only once; inline its value.">$copy</warning> = load();
            return $copy;
        };
        return $this->slots;
    }
}
