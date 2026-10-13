<?php
$hex = <weak_warning descr="Sort the integer keys ascending so the array can be stored packed.">[</weak_warning>0x1 => 'a', 0 => 'b', 2 => 'c'];
$bits = <weak_warning descr="Sort the integer keys ascending so the array can be stored packed.">[</weak_warning>0b11 => 'a', 0o1 => 'b', 1_000 => 'c'];
$big = <weak_warning descr="Sort the integer keys ascending so the array can be stored packed.">[</weak_warning>5000000000 => 'a', 4000000000 => 'b', 6000000000 => 'c'];
$bigStr = <weak_warning descr="Write the keys as integers so the array can be stored packed.">[</weak_warning>'5000000000' => 'a', 6000000000 => 'b', 7000000000 => 'c'];

$sortedHex = [0x1 => 'a', 0x2 => 'b', 010 => 'c'];
$tooBig = ['9223372036854775808' => 'a', 2 => 'b', 1 => 'c'];
$floatish = [9223372036854775808 => 'a', 2 => 'b', 1 => 'c'];
$mixedBig = <weak_warning descr="Sort the integer keys ascending so the array can be stored packed.">[</weak_warning>3 => 'a', 2 => 'b', 4294967296 => 'c'];
