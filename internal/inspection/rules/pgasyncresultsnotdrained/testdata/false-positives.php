<?php
function f($c) { if(pg_send_query($c,'SELECT 7; SELECT 8')){ while(pg_get_result($c)!==false){} pg_send_query($c,'SELECT 9'); } }
