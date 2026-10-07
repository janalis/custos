<?php
class Config { public static $label = ''; const NAME = 'x'; }
function pick() {
    $a = Config::$label ?: 'none';
    $b = Config::NAME ?: 'none';
}
