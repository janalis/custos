<?php
class Registry { public static array $items = []; }
function add($v) {
    Registry::$items[<warning descr="The index is redundant here; use '[]' to append.">count</warning>(registry::$items)] = $v;
}
