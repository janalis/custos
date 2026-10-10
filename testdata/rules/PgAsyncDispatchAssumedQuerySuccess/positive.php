<?php
function f($c) { if(<warning descr="Inspect the asynchronous PostgreSQL result status.">pg_send_query($c,'SELECT 7')</warning>){ return true; } return false; }
