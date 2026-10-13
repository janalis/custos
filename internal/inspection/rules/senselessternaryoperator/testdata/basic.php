<?php
$a = <warning descr="This ternary always yields '$right'; use it directly.">$left === $right ? $left : $right</warning>;
$b = <warning descr="This ternary always yields '$left'; use it directly.">$left !== $right ? $left : $right</warning>;
$c = <warning descr="This ternary always yields '$size'; use it directly.">$size === -1 ? -1 : $size</warning>;
$d = <warning descr="This ternary always yields 'null'; use it directly.">($ref === null) ? $ref : null</warning>;
$e = <warning descr="This ternary always yields '$t'; use it directly.">$t !== 'n/a' ? $t : 'n/a'</warning>;
$f = f(<warning descr="This ternary always yields '1.5'; use it directly.">$ratio !== 1.5 ? 1.5 : $ratio</warning>);
$k = (<warning descr="This ternary always yields '$o -> p'; use it directly.">$o->p === $q ? $q : $o -> p</warning>);
