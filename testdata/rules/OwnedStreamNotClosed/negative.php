<?php
function good($path) { $h = fopen($path, 'rb'); if ($h === false) { return false; } try { return fgets($h); } finally { fclose($h); } }
