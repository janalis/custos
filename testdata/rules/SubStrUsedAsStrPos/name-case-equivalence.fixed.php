<?php
class Prefix { public static string $value = 'x'; }
function starts($s) {
    return strpos($s, prefix::$value) === 0;
}
