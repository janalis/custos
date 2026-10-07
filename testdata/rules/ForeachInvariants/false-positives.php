<?php
function f(array $a, array $b, $flag) {
    for ($i = 0; $i < count($a); $i += 1) { echo $a[$i]; }
    for ($i = 0; $i < count($a); $i--) { echo $a[$i]; }
    for ($i = 0.0; $i < count($a); $i++) { echo $a[$i]; }
    for ($i = 0; $i < count($a) || $stop; $i++) { echo $a[$i]; }
    for ($i = 0; $i < count($a); $i++) { echo $a[$i], $b[$i]; }
    for ($i = 0; $i < sizeof($a); $i++) { echo $a[$i]; }
    for ($i = 0; $i < count($a); $i++) {}
    for ($i = 0; $i < count($a); $i++) echo $a[$i];
    for ($i = 0; $i < count($a, COUNT_RECURSIVE); $i++) { echo $a[$i]; }
    $k = count($a);
    if ($flag) {
        $k = count($b);
    }
    for ($i = 0; $i < $k; $i++) { echo $a[$i]; }
}
$m = count($top);
for ($i = 0; $i < $m; $i++) { echo $top[$i]; }
