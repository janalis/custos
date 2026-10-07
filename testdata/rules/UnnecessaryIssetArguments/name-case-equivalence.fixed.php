<?php
class Cache { public static $data = []; }
function has() {
    return isset(cache::$data['k']);
}
