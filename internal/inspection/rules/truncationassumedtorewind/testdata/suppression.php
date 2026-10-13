<?php
// @custos-ignore TruncationAssumedToRewind
$fp = fopen($path, "w+"); if ($fp === false) { return; } if (fseek($fp, 10) !== 0) { return; } if (ftruncate($fp, 0)) { fwrite($fp, "new"); }
