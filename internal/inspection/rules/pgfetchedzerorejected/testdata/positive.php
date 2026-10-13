<?php
function f($r) { if(<warning descr="Distinguish PostgreSQL fetch failure from a zero value.">$v=pg_fetch_result($r,0,0)</warning>){echo $v;} }
function bare($r) { if (<warning descr="Distinguish PostgreSQL fetch failure from a zero value.">pg_fetch_result($r, 0, 0)</warning>) { echo 'found'; } }
function loop($r) { $i = 0; while (<warning descr="Distinguish PostgreSQL fetch failure from a zero value.">$v = pg_fetch_result($r, $i++, 0)</warning>) { echo $v; } }
function commented($r) { if (<warning descr="Distinguish PostgreSQL fetch failure from a zero value.">$v = pg_fetch_result($r, /* row */ 0, 0)</warning>) { echo $v; } }
function qualified($r) { if (<warning descr="Distinguish PostgreSQL fetch failure from a zero value.">\pg_fetch_result($r, 0, 'is_admin')</warning>) { echo 'admin'; } }
