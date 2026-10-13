<?php
class Cache { public static $map = []; }
function lookup() {
    return <weak_warning descr="Simplify to 'cache::$map['k'] ?? 'none'' using the null coalescing operator.">isset(Cache::$map['k']) ? cache::$map['k'] : 'none'</weak_warning>;
}
