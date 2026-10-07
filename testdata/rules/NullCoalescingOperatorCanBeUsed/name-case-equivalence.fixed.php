<?php
class Cache { public static $map = []; }
function lookup() {
    return cache::$map['k'] ?? 'none';
}
