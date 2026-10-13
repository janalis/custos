<?php
function bad($privateDir, $secret) { $path = tempnam($privateDir, 'job'); if ($path === false) { return; } <warning descr="Verify the temporary file remains in the required directory.">file_put_contents($path, $secret)</warning>; }
