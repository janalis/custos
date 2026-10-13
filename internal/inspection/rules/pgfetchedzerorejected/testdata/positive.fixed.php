<?php
function f($r) { if(is_string($v=pg_fetch_result($r,0,0))){echo $v;} }
function bare($r) { if (is_string(pg_fetch_result($r, 0, 0))) { echo 'found'; } }
function loop($r) { $i = 0; while (is_string($v = pg_fetch_result($r, $i++, 0))) { echo $v; } }
function commented($r) { if (is_string($v = pg_fetch_result($r, /* row */ 0, 0))) { echo $v; } }
function qualified($r) { if (is_string(\pg_fetch_result($r, 0, 'is_admin'))) { echo 'admin'; } }
