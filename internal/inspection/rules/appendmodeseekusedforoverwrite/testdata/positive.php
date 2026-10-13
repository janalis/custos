<?php
$fp = fopen($path, "a"); if ($fp === false) { return; } if (fseek($fp, 0) !== 0) { return; } <warning descr="Use an overwrite-capable stream mode.">fwrite($fp, "head")</warning>;
