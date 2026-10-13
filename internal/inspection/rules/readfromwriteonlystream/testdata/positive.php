<?php
$fp = fopen($path, "w"); if ($fp === false) { return; } <error descr="Open the stream with read access.">fread($fp, 8)</error>;
