<?php
$double = function ($p) { $r = []; <weak_warning descr="Use a braced block for the body of this construct.">foreach</weak_warning> ([0, 1] as $k) $r[$k] = $p[$k] * 2; return $r; };
foreach ($rows as $e) { <weak_warning descr="Use a braced block for the body of this construct.">if</weak_warning> (!isset($e['lp'])) continue; echo $e['lp']; }
<weak_warning descr="Use a braced block for the body of this construct.">if</weak_warning> ($a) echo 1; // note
<weak_warning descr="Use a braced block for the body of this construct.">if</weak_warning> ($b) echo 2; /* note */
