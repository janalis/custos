<?php
function bad($path) { $h = <warning descr="Acquire the lock before truncating the file.">fopen($path, 'wb')</warning>; if ($h !== false) { flock($h, LOCK_EX); fclose($h); } }
function badNonBlocking($path) { $h = <warning descr="Acquire the lock before truncating the file.">fopen($path, 'w+')</warning>; if ($h !== false && flock($h, LOCK_EX | LOCK_NB)) { fwrite($h, 'x'); } }
function badAfterUnlock($path) { $h = <warning descr="Acquire the lock before truncating the file.">fopen($path, 'w')</warning>; flock($h, LOCK_UN); flock(operation: LOCK_EX, stream: $h); fclose($h); }
