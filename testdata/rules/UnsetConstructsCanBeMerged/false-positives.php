<?php
unset($a);
;
unset($b);
$z = 0;
unset($c["x"], $d);
$e = 1;
unset($f);
if ($g) {
    unset($h);
}
unset($i);
switch ($j) {
    case 1:
        unset($k);
        break;
    case 2:
        unset($l);
}

if ($drop) unset($cache);
