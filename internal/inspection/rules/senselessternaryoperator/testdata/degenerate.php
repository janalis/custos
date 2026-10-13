<?php
$a = $x === $x ? $x : $y;
$b = $x !== $x ? $y : $x;
$c = ($v === $v) ? $v : null;
$d = $n === 1 ? 1 : 2;
$e = <warning descr="This ternary always yields '$y'; use it directly.">$y === $x ? $x : $y</warning>;
$f = <warning descr="This ternary always yields '$x'; use it directly.">$x === $x ? $x : $x</warning>;
