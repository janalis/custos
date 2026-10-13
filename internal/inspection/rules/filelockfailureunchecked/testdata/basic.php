<?php
function bad($path) { $h = fopen($path, 'cb'); if ($h === false) { return; } flock($h, LOCK_EX); <warning descr="Acquire the lock successfully before writing.">fwrite($h, 'next')</warning>; fclose($h); }
