<?php
function unstarted_directory_generator(string $path) {
    $entries = scandir($path);
    if ($entries === false) { return; }
    $replace = function () use (&$entries, $path) { $entries = scandir($path); yield 1; };
    $replace();
    foreach ($entries as $entry) { echo $entry; }
}
