<?php
foreach (new DirectoryIterator('/tmp') as $entry) { $entry = new SplFileInfo('/safe/file'); unlink($entry->getPathname()); }
foreach (new DirectoryIterator('/tmp') as $entry) { if ($flag) { $entry = new SplFileInfo('/safe/file'); } unlink($entry->getPathname()); }
