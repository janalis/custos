<?php
function f($c) { if(!pg_send_query($c,'SELECT 7')){return false;} $r=pg_get_result($c); return $r!==false && pg_result_status($r)===PGSQL_TUPLES_OK; }
