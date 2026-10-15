<?php
$entries = scandir($directory);
foreach (<warning descr="Check directory enumeration before iterating.">$entries</warning> as $entry) { echo $entry; }

$checked = scandir($directory);
if ($checked === false) { exit; }
foreach ($checked as $entry) { echo $entry; }
