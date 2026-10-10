<?php
if (!$data = <warning descr="Compare socket read failure strictly with false.">socket_read($s, 1024, PHP_BINARY_READ)</warning>) { throw new RuntimeException(); }
