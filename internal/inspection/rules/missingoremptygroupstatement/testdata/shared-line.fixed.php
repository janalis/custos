<?php
$double = function ($p) { $r = []; foreach ([0, 1] as $k) { $r[$k] = $p[$k] * 2; } return $r; };
foreach ($rows as $e) { if (!isset($e['lp'])) { continue; } echo $e['lp']; }
if ($a) {
    echo 1;
} // note
if ($b) {
    echo 2;
} /* note */
