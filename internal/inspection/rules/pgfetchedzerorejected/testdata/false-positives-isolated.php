<?php
function f($r) { if(($v=pg_fetch_result($r,0,0))!==false){echo $v;} }
function strings($r) { if (is_string($v = pg_fetch_result($r, 0, 0))) { echo $v; } }
function bare($r) { if (is_string(pg_fetch_result($r, 0, 0))) { echo 'found'; } }
