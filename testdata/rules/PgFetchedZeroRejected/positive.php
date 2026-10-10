<?php
function f($r) { if(<warning descr="Distinguish PostgreSQL fetch failure from a zero value.">$v=pg_fetch_result($r,0,0)</warning>){echo $v;} }
