<?php
if (is_file($_SERVER['SCRIPT_FILENAME'])) {
    return false;
}
if ($legacy) {
    return @(require __DIR__ . '/legacy.php');
}
return require __DIR__ . '/index.php';
return $cached ?: <weak_warning descr="Avoid relying on the value returned by an included file.">include 'fresh.php'</weak_warning>;
return $ok || <weak_warning descr="Avoid relying on the value returned by an included file.">require 'other.php'</weak_warning>;

function load()
{
    return <weak_warning descr="Avoid relying on the value returned by an included file.">require __DIR__ . '/config.php'</weak_warning>;
}

$load = fn () => <weak_warning descr="Avoid relying on the value returned by an included file.">include 'lazy.php'</weak_warning>;
