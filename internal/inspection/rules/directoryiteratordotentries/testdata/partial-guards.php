<?php
function oneName($dir) { foreach (new DirectoryIterator($dir) as $e) { if ($e->getFilename() === '.') { continue; } <warning descr="Skip dot entries before modifying directory entries.">rmdir($e->getPathname())</warning>; } }
function optionalGuard($dir, $force) { foreach (new DirectoryIterator($dir) as $e) { if (!$e->isDot() || $force) { <warning descr="Skip dot entries before modifying directory entries.">rmdir($e->getPathname())</warning>; } } }
function directoriesOnly($dir) { foreach (new DirectoryIterator($dir) as $e) { if ($e->isDir()) { <warning descr="Skip dot entries before modifying directory entries.">rmdir($e->getPathname())</warning>; } } }
