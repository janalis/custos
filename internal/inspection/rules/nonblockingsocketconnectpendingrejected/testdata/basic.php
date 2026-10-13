<?php
if (socket_set_nonblock($s)) { if (!<warning descr="Handle a pending nonblocking connection.">socket_connect($s, $host, 443)</warning>) { throw new RuntimeException(); } }
