<?php
$p = [1, 2];
<weak_warning descr="Use short destructuring syntax '[...] = ...'.">list</weak_warning> /* pair */ ($a, $b) = $p;
<weak_warning descr="Use short destructuring syntax '[...] = ...'.">list</weak_warning> // x
($c) = $p;
<weak_warning descr="Use short destructuring syntax '[...] = ...'.">list</weak_warning>/* glued */($d, list /* inner */ ($e)) = [1, [2]];
foreach ([$p] as <weak_warning descr="Use short destructuring syntax in foreach ('as [...]').">list</weak_warning> # hash
    ($f, $g)) {
    echo $f, $g;
}
echo $a, $b, $c, $d, $e, "\n";
