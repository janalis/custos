<?php
// Forwarded arguments that are not provably one positional value: reported, never fixed.
function labels(array $l, $cb = 'uniqid') { return array_map(<error descr="Pass more_entropy = true to uniqid() to reduce collisions.">$cb</error>, $l); }
function local(array $l) { $cb = 'uniqid'; return array_map(<error descr="Pass more_entropy = true to uniqid() to reduce collisions.">$cb</error>, $l); }
function ids(array $args) { return call_user_func_array(<error descr="Pass more_entropy = true to uniqid() to reduce collisions.">'uniqid'</error>, $args); }
$a = call_user_func_array(<error descr="Pass more_entropy = true to uniqid() to reduce collisions.">'uniqid'</error>, []);
$b = call_user_func_array(<error descr="Pass more_entropy = true to uniqid() to reduce collisions.">'uniqid'</error>, ['prefix' => 'x']);
$c = call_user_func_array(<error descr="Pass more_entropy = true to uniqid() to reduce collisions.">'uniqid'</error>, [...$parts]);
$d = call_user_func_array(<error descr="Pass more_entropy = true to uniqid() to reduce collisions.">'uniqid'</error>, [&$ref]);
$e = call_user_func(<error descr="Pass more_entropy = true to uniqid() to reduce collisions.">'uniqid'</error>, prefix: 'x');
$f = call_user_func(<error descr="Pass more_entropy = true to uniqid() to reduce collisions.">'uniqid'</error>, ...$parts);
$g = array_map(array: $l, callback: <error descr="Pass more_entropy = true to uniqid() to reduce collisions.">'uniqid'</error>);
$h = call_user_func_array(args: ['x'], callback: <error descr="Pass more_entropy = true to uniqid() to reduce collisions.">'uniqid'</error>);
$i = array_map(<error descr="Pass more_entropy = true to uniqid() to reduce collisions.">'uniqid'</error>, more_entropy: $l);

// Comments, spacing and trailing commas inside the parentheses survive F1.
$j = <error descr="Pass more_entropy = true to uniqid() to reduce collisions.">uniqid(/* empty on purpose */)</error>;
$k = <error descr="Pass more_entropy = true to uniqid() to reduce collisions.">uniqid('a' /* prefix */ ,)</error>;
$m = <error descr="Pass more_entropy = true to uniqid() to reduce collisions.">uniqid(prefix: 'q',)</error>;
$n = <error descr="Pass more_entropy = true to uniqid() to reduce collisions.">uniqid   ( $p )</error>;
