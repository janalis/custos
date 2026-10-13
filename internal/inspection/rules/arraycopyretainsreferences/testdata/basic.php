<?php
$a=[1]; $r=&$a[0]; $b=$a; <warning descr="Detach referenced array elements before modifying the copy.">$b[0]=9</warning>; echo $a[0];
