<?php
$fp = fopen("php://memory", "r+"); fwrite($fp, "
"); rewind($fp); if (<warning descr="Recognize a blank CSV record as a null field.">fgetcsv($fp) === []</warning>) {}
