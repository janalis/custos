<?php
// @custos-ignore AppendModeSeekUsedForOverwrite
$fp = fopen($path, "a"); if ($fp === false) { return; } if (fseek($fp, 0) !== 0) { return; } fwrite($fp, "head");
