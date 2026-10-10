<?php
$data = socket_read($s, 1024, PHP_BINARY_READ); if ($data === false) { throw new RuntimeException(); }
