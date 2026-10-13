<?php
$n = 0; $next = <warning descr="Use persistent shared state for this counter.">fn() => ++$n</warning>; echo $next(), $next();
