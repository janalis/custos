<?php
class Prefix { public static function get() { return 'x'; } }
function starts($s) {
    return <weak_warning descr="Use 'strpos($s, prefix::get()) === 0' instead.">substr($s, 0, strlen(Prefix::Get())) === prefix::get()</weak_warning>;
}
