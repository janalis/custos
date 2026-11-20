<?php // presets: per laravel
foo(... $args);
foo(name : $v, other:1);
foo(&$legacy);
$a = [ ...$b, &$c, 'k' => &$d ];
if (isset( $a, $b ) || empty( $c )) { eval( 'x' ); }
$f = strlen( ... );
#[Route( '/x', methods: [ 'GET' ] )]
function g(#[SensitiveParameter] $p, #[A] #[B] int $q) {}
$c = #[Pure] fn() => 1;
$d = #[Pure] static function () {};
exit;
die('x');
$x = clone($y);
