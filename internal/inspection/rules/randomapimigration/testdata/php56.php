<?php
$a = <warning descr="Prefer mt_rand() over rand().">rand(1, 2)</warning>;
$b = mt_rand(1, 2);
