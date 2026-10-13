<?php
class Prefix { public static string $value = 'x'; }
function starts($s) {
    return <weak_warning descr="Use 'strpos($s, prefix::$value) === 0' instead.">substr($s, 0, strlen(Prefix::$value)) === prefix::$value</weak_warning>;
}
