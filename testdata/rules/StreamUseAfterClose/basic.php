<?php
function bad($path) { $h = fopen($path, 'wb'); if ($h === false) { return; } fclose($h); <error descr="Use an open stream for this operation.">fwrite($h, 'next')</error>; }
