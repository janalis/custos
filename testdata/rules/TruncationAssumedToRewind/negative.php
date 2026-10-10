<?php
$fp = fopen($path, "w+"); if ($fp === false) { return; } if (fseek($fp, 10) !== 0) { return; } if (ftruncate($fp, 0)) { rewind($fp); fwrite($fp, "new"); }
