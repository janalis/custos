<?php
function bad($dir) { foreach (new DirectoryIterator($dir) as $entry) { <warning descr="Skip dot entries before modifying directory entries.">unlink($entry->getPathname())</warning>; } }
