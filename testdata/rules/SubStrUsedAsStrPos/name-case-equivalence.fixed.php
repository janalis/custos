<?php
class Prefix { public static function get() { return 'x'; } }
function starts($s) {
    return strpos($s, prefix::get()) === 0;
}
