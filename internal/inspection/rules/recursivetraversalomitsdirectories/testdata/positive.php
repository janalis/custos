<?php
$dir = new RecursiveDirectoryIterator("."); $it = <warning descr="Include parent nodes when traversing directories.">new RecursiveIteratorIterator($dir)</warning>;
