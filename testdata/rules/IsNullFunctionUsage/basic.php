<?php

$a = <weak_warning descr="Replace with '$row === null'.">is_null($row)</weak_warning>;
$b = <weak_warning descr="Replace with '$row !== null'.">!is_null($row)</weak_warning>;
$c = <weak_warning descr="Replace with '$cfg->get() === null'.">is_null($cfg->get()) == TRUE</weak_warning>;
$d = <weak_warning descr="Replace with '$cfg->get() !== null'.">false === is_null($cfg->get())</weak_warning>;
$e = <weak_warning descr="Replace with '$row === null'.">is_null($row) != false</weak_warning>;
$f = <weak_warning descr="Replace with '$row !== null'.">\is_null($row) !== true</weak_warning>;
$g = $ok and <weak_warning descr="Replace with '$row === null'.">is_null($row)</weak_warning>;
$h = <weak_warning descr="Replace with '($left ?? $right) !== null'.">!is_null($left ?? $right)</weak_warning>;
$i = <weak_warning descr="Replace with '($n = next($it)) === null'.">is_null($n = next($it))</weak_warning>;
$j = <weak_warning descr="Replace with '(int) $raw === null'.">is_null((int) $raw)</weak_warning>;
$l = !(<weak_warning descr="Replace with '$row === null'.">is_null($row)</weak_warning>);
$m = <weak_warning descr="Replace with '$row === null'.">is_null($row)</weak_warning> && $other;
$k = is_null($p, $q);
$o = $x->is_null($row);
