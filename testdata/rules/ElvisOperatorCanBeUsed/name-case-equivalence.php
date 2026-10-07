<?php
class Config { public static $label = ''; const NAME = 'x'; }
function pick() {
    $a = <weak_warning descr="Use the short ternary: 'Config::$label ?: 'none''.">Config::$label ? config::$label : 'none'</weak_warning>;
    $b = <weak_warning descr="Use the short ternary: 'Config::NAME ?: 'none''.">Config::NAME ? CONFIG::NAME : 'none'</weak_warning>;
}
