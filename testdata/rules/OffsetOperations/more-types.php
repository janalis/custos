<?php
class KeyMap {
    public function offsetGet(BaseKey $k) { return null; }
}
class BaseKey {}
class ObjMap {
    public function offsetGet(object $k) { return null; }
}

/**
 * @param array|false $af
 * @param callable|array $ca
 * @param null|object $no
 */
function more(bool $flag, Closure $fn, callable $cb, $af, $ca, $no, KeyMap $map) {
    echo <error descr="'$flag' does not support offset access (types: bool).">$flag[0]</error>;
    echo $fn[0];
    echo $cb[0];
    echo $af[0];
    echo $ca[0];
    echo $no[0];
    echo $map[<error descr="Index of type int does not fit the accepted \BaseKey.">1</error>];
}

function objects(ObjMap $m) {
    echo $m[new DateTime()];
    echo $m[<error descr="Index of type int does not fit the accepted object.">1</error>];
}

function late() {
    echo <error descr="'$x' does not support offset access (types: int).">$x[0]</error>;
    $x = 5;
}
