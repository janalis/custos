<?php

echo 1; // presets: per laravel
$a = $b
    && $c;
$d = $e ||
    $f;
$g = 'x'.
    'y';
declare(strict_types=1);
static $s = 1;
global $gl;
$fn = function () use (&$a): void {};
$q = (object) ['a' => 1];
$r = &$q;
$t = $obj->{'prop'};
$u = ${'var'};
if ($a) foo();
elseif ($b) bar();
else baz();
while ($a)
    $a--;
label:
$v = yield $k => $w;
$y = clone $x;
print $z;
exit(0);
