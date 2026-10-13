<?php
// @custos-ignore WriteToReadOnlyStream
$fp = fopen($path, "r"); if ($fp === false) { return; } fwrite($fp, "change");
