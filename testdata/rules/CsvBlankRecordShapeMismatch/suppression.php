<?php
// @custos-ignore CsvBlankRecordShapeMismatch
$fp = fopen("php://memory", "r+"); fwrite($fp, "
"); rewind($fp); if (fgetcsv($fp) === []) {}
