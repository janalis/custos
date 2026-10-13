<?php
function interpolation(array $a, array $o) {
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < count($a); $i++) {
        echo "{$a[$i]}abc {$a[$i]} $a[$i] $a[$i]_x $a[$i][0] {$a[$i]}->p $a[$i]-x é{$a[$i]}é";
        $a[$i]++;
    }
}
