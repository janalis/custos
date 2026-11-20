<?php
// presets: per; set: array_syntax=long, list_syntax=long, new_parentheses=remove
$a = array(1, array(2));
list($x, $y) = $a;
$o = new Foo;
$p = new Foo()->bar();
$q = (new Foo)->bar();
$r = new class () {};
$s = new Foo(/* c */);
