<?php
function good($path) { $h = fopen($path, 'rb'); if ($h === false) { return; } fread($h, 12); fclose($h); }
