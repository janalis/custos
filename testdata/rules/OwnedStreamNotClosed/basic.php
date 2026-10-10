<?php
function bad($path) { $h = fopen($path, 'rb'); if ($h === false) { return false; } <warning descr="Close the owned stream before leaving its scope.">return fgets($h);</warning> }
