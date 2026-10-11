<?php
// Forwarded arguments that are not provably one positional value: reported, never fixed.
function labels(array $l, $cb = 'uniqid') { return array_map($cb, $l); }
function local(array $l) { $cb = 'uniqid'; return array_map($cb, $l); }
function ids(array $args) { return call_user_func_array('uniqid', $args); }
$a = call_user_func_array('uniqid', []);
$b = call_user_func_array('uniqid', ['prefix' => 'x']);
$c = call_user_func_array('uniqid', [...$parts]);
$d = call_user_func_array('uniqid', [&$ref]);
$e = call_user_func('uniqid', prefix: 'x');
$f = call_user_func('uniqid', ...$parts);
$g = array_map(array: $l, callback: 'uniqid');
$h = call_user_func_array(args: ['x'], callback: 'uniqid');
$i = array_map('uniqid', more_entropy: $l);

// Comments, spacing and trailing commas inside the parentheses survive F1.
$j = uniqid('', true/* empty on purpose */);
$k = uniqid('a', true /* prefix */ ,);
$m = uniqid(prefix: 'q', more_entropy: true,);
$n = uniqid   ( $p, true );
