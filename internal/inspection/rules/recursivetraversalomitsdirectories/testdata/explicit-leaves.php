<?php
$dir = new RecursiveDirectoryIterator(".");
$files = new RecursiveIteratorIterator($dir, RecursiveIteratorIterator::LEAVES_ONLY);
$also = new RecursiveIteratorIterator($dir, 0);
$named = new RecursiveIteratorIterator(mode: RecursiveIteratorIterator::LEAVES_ONLY, iterator: $dir);
$omitted = <warning descr="Include parent nodes when traversing directories.">new RecursiveIteratorIterator($dir)</warning>;
