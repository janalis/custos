<?php
function f($c) { if(pg_send_query($c,'SELECT 7; SELECT 8')){ pg_get_result($c); <warning descr="Drain all pending PostgreSQL results before another asynchronous query.">pg_send_query($c,'SELECT 9')</warning>; } }
