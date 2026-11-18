<?php
function good($dir) { foreach (new DirectoryIterator($dir) as $entry) { if ($entry->isDot()) { continue; } unlink($entry->getPathname()); } }
