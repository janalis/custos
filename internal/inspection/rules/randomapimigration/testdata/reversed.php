<?php
$a = <warning descr="Prefer mt_rand() over rand().">rand(10, 1)</warning>;
$b = <warning descr="Prefer mt_rand() over rand().">rand(5, -5)</warning>;
$c = <warning descr="Prefer mt_rand() over rand().">rand(max: 1, min: 10)</warning>;
$d = <warning descr="Prefer mt_rand() over rand().">rand(max: 10, min: 1)</warning>;
$e = <warning descr="Prefer mt_rand() over rand().">rand(1, 10)</warning>;
$f = <warning descr="Prefer mt_rand() over rand().">rand(3, 3)</warning>;
$g = <warning descr="Prefer mt_rand() over rand().">rand(10, $hi)</warning>;
$h = <warning descr="Prefer mt_rand() over rand().">rand($lo, 1)</warning>;
$i = <warning descr="Prefer mt_rand() over rand().">rand(10, max: 1)</warning>;
