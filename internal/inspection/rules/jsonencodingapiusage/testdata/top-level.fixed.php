<?php
// A script's own code is one scope: its first read after the decode counts.
$config = json_decode((string) file_get_contents('config.json'), true);
if (!$config) {
    exit(1);
}
foreach (glob('*.json') as $file) {
    $row = json_decode((string) file_get_contents($file), true);
    if (!$row) {
        continue;
    }
    echo $row['id'];
}
$raw = json_decode('[]', true, 512, JSON_THROW_ON_ERROR);
function later($raw) {
    return $raw ?: [];
}
echo $raw['a'];
