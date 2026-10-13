<?php
interface Key {}
class BaseKey implements Key {}
class SpecialKey extends BaseKey {}
class OtherKey {}

class KeyMap {
    public function offsetGet(BaseKey $k) { return null; }
}
class KeyIfaceMap {
    public function offsetGet(Key $k) { return null; }
}

function lookup(KeyMap $map, KeyIfaceMap $imap, SpecialKey $special, BaseKey $base, OtherKey $other) {
    echo $map[$base];
    echo $map[$special];
    echo $imap[$special];
    echo $map[<error descr="Index of type \OtherKey does not fit the accepted \BaseKey.">$other</error>];
    echo $imap[<error descr="Index of type \OtherKey does not fit the accepted \Key.">$other</error>];
}
