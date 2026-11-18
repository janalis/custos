<?php
$fp = fopen($path, "r+"); if ($fp === false) { return; } if (fseek($fp, 0) !== 0) { return; } fwrite($fp, "head");
