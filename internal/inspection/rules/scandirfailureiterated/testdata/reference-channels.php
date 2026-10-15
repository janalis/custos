<?php
function captured_directory(string $first, string $next) {
    $entries = scandir($first);
    if ($entries === false) { return; }
    $replace = function () use (&$entries, $next) { $entries = scandir($next); };
    $replace();
    foreach (<warning descr="Check directory enumeration before iterating.">$entries</warning> as $entry) { echo $entry; }
}
function checked_directory_replacement(string $first, string $next) {
    $entries = scandir($first);
    if ($entries === false) { return; }
    $replace = function () use (&$entries, $next) { $entries = scandir($next); };
    $replace();
    if ($entries === false) { return; }
    foreach ($entries as $entry) { echo $entry; }
}
function overwritten_directory_replacement(string $first, string $next) {
    $entries = scandir($first);
    if ($entries === false) { return; }
    $replace = function () use (&$entries, $next) { $entries = scandir($next); };
    $replace();
    $entries = [];
    foreach ($entries as $entry) { echo $entry; }
}
final class DirectoryWriter {
    public function replace(&$entries, string $path) { $entries = scandir($path); }
}
function directory_reference_method(DirectoryWriter $writer, string $path) {
    $entries = [];
    $writer->replace($entries, $path);
    foreach (<warning descr="Check directory enumeration before iterating.">$entries</warning> as $entry) { echo $entry; }
}
function directory_global_writer(string $path) { global $directory_entries; $directory_entries = scandir($path); }
function directory_global_reader(string $path) {
    global $directory_entries;
    $directory_entries = [];
    directory_global_writer($path);
    foreach (<warning descr="Check directory enumeration before iterating.">$directory_entries</warning> as $entry) { echo $entry; }
}
