<?php
function f($c) { if(!pg_send_query($c,'SELECT 7')){return false;} $r=pg_get_result($c); return $r!==false && pg_result_status($r)===PGSQL_TUPLES_OK; }
function g($c) { if (pg_send_query($c, 'SELECT 8')) return false; return true; }
function h($c) { if (pg_send_query($c, 'SELECT 9')) return pg_get_result($c); return false; }
function i($c) { if (pg_send_query($c, 'SELECT 10')) { } return false; }
