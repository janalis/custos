<?php
function good($path) { $h = fopen($path, 'wb'); if ($h === false) { return; } fwrite($h, 'next'); fclose($h); }
