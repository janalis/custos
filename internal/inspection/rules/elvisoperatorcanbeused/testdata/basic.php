<?php
$label  = <weak_warning descr="Use the short ternary: '$name ?: 'guest''.">$name ? $name : 'guest'</weak_warning>;
$limit  = <weak_warning descr="Use the short ternary: '$cfg['max'] ?: 10'.">$cfg['max'] ? ( $cfg['max'] ) : 10</weak_warning>;
$owner  = <weak_warning descr="Use the short ternary: '(($user)) ?: $fallback'.">(($user)) ? $user : $fallback</weak_warning>;
$title  = <weak_warning descr="Use the short ternary: '$page->title ?: (DEFAULT_TITLE)'.">$page->title ? $page->title : (DEFAULT_TITLE)</weak_warning>;
$const  = <weak_warning descr="Use the short ternary: 'Config::$name ?: f()'.">Config::$name ? Config::$name : f()</weak_warning>;
$wrap   = (<weak_warning descr="Use the short ternary: '$a->b ?: 0'.">$a->b ? $a->b : 0</weak_warning>) + 1;
