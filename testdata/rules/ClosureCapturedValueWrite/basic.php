<?php
$n = 0; $f = <warning descr="Capture the scalar by reference when updating outer state.">function () use ($n) { ++$n; }</warning>; $f(); echo $n;
