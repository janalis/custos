<?php
$p = [1, 2];
/* pair */ [$a, $b] = $p;
// x
[$c] = $p;
/* glued */[$d, /* inner */ [$e]] = [1, [2]];
foreach ([$p] as # hash
    [$f, $g]) {
    echo $f, $g;
}
echo $a, $b, $c, $d, $e, "\n";
