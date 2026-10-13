<?php
class Cache { public static $data = []; }
function has() {
    return isset(<weak_warning descr="Redundant isset() argument: a deeper array access already covers it.">Cache::$data</weak_warning>, cache::$data['k']);
}
