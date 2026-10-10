<?php
class Registry { public static array $items = []; }
function add($v) {
    Registry::$items[count(registry::$items)] = $v;
}
