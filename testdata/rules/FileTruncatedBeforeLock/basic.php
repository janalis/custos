<?php
function bad($path) { $h = <warning descr="Acquire the lock before truncating the file.">fopen($path, 'wb')</warning>; if ($h !== false) { flock($h, LOCK_EX); fclose($h); } }
