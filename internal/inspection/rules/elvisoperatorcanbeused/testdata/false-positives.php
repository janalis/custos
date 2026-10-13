<?php
$keep1  = $name ?: 'guest';
$keep2  = $name ? 'guest' : $name;
$keep3  = !$name ? $name : 'guest';
$keep4  = isset($name) ? $name : 'guest';
$keep5  = $name ? $other : 'guest';
$keep6  = $cfg['max'] ? $cfg['min'] : 10;
