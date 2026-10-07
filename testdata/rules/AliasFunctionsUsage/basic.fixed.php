<?php
$total = count($rows);
$csv   = \implode(';', $cells);
$ok    = is_int($port) && is_writable($dir);
magic_quotes_runtime(0);
