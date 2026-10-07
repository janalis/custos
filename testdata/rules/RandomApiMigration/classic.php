<?php
<warning descr="Prefer mt_srand() over srand().">srand(1234)</warning>;
$cap  = <warning descr="Prefer mt_getrandmax() over getrandmax().">getrandmax()</warning>;
$roll = <warning descr="Prefer mt_rand() over rand().">\rand(1, 6)</warning>;
$keep = mt_rand(1, 6);
