<?php
$fp = fopen($path, "w+"); if ($fp === false) { return; } if (fseek($fp, 10) !== 0) { return; } if (ftruncate($fp, 0)) { <warning descr="Rewind after truncating the stream.">fwrite($fp, "new")</warning>; }
