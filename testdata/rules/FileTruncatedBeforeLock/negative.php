<?php
function good($path) { $h = fopen($path, 'cb'); if ($h === false) { return; } if (flock($h, LOCK_EX)) { ftruncate($h, 0); } fclose($h); }
