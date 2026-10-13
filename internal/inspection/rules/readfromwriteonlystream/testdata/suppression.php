<?php
// @custos-ignore ReadFromWriteOnlyStream
$fp = fopen($path, "w"); if ($fp === false) { return; } fread($fp, 8);
