<?php
function good($path) { $h = fopen($path, 'cb'); if ($h === false) { return; } if (flock($h, LOCK_EX)) { fwrite($h, 'next'); flock($h, LOCK_UN); } fclose($h); }
