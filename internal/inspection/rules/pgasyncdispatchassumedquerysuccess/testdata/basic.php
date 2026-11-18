<?php
function f($c) { if(<warning descr="Inspect the asynchronous PostgreSQL result status.">pg_send_query($c,'SELECT 7')</warning>){ return true; } return false; }
function g($c) { if (<warning descr="Inspect the asynchronous PostgreSQL result status.">pg_send_query($c, 'SELECT 8')</warning>) return true; return false; }
function h($c) { if (<warning descr="Inspect the asynchronous PostgreSQL result status.">pg_send_query($c, 'SELECT 9')</warning>): return true; endif; return false; }
function i($c) { if (<warning descr="Inspect the asynchronous PostgreSQL result status.">pg_send_query($c, 'SELECT 10')</warning>) echo 'sent'; }
